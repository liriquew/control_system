package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	authclient "github.com/liriquew/tasks_service/internal/grpc/clients/auth_client"
	grphclient "github.com/liriquew/tasks_service/internal/grpc/clients/graphs_client"
	predictionsclient "github.com/liriquew/tasks_service/internal/grpc/clients/predictions_client"
	"github.com/liriquew/tasks_service/internal/models"
	repository "github.com/liriquew/tasks_service/internal/repository"
	"github.com/liriquew/tasks_service/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type tasksRepository interface {
	SaveTask(ctx context.Context, task *tsks_pb.Task) (int64, error)
	GetGroupTasks(ctx context.Context, taskID int64) ([]*models.Task, error)
	GetTaskByID(ctx context.Context, taskID int64) (*models.Task, error)
	GetTaskList(ctx context.Context, userID, offset int64) ([]*models.Task, error)
	UpdateTask(ctx context.Context, task *tsks_pb.Task) error
	UpdateGroupTask(ctx context.Context, task *tsks_pb.Task) error
	DeleteUserTask(ctx context.Context, userID, taskID int64) error
	DeleteGroupTask(ctx context.Context, taskID int64) error

	TaskInAnyGroup(ctx context.Context, taskID int64) (int64, error)
	TaskInGroup(ctx context.Context, groupID, taskID int64) error
	GetTasks(ctx context.Context, taskIDs []int64) ([]*models.Task, error)
}

type tasksProducer interface {
	ProduceTaskPredictionData(context.Context, *models.TaskPredictionData) error
	ProduceTaskPredictionDataDelete(context.Context, int64) error
}

type serverAPI struct {
	tsks_pb.UnimplementedTasksServer
	authClient *authclient.AuthClient
	prdtClient *predictionsclient.PredicionsClient
	grphClient *grphclient.GraphClient
	repository tasksRepository
	producer   tasksProducer
	log        *slog.Logger
}

func Register(gRPC *grpc.Server, taskServiceAPI tsks_pb.TasksServer) {
	tsks_pb.RegisterTasksServer(gRPC, taskServiceAPI)
}

func NewServerAPI(log *slog.Logger,
	taskRepository tasksRepository,
	producer tasksProducer,
	authClient *authclient.AuthClient,
	prdtClient *predictionsclient.PredicionsClient,
	grphClient *grphclient.GraphClient,
) *serverAPI {
	return &serverAPI{
		log:        log,
		repository: taskRepository,
		producer:   producer,
		authClient: authClient,
		prdtClient: prdtClient,
		grphClient: grphClient,
	}
}

var (
	ErrMissingJWT = errors.New("miss jwt token")
	ErrDeny       = errors.New("denied")
)

func (s *serverAPI) authenticate(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.log.Error("error while extracting metadata")
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	AuthParams := md.Get("user-id")
	if len(AuthParams) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user-idmetadata")
	}
	userID, err := strconv.ParseInt(AuthParams[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid user-id metadata")
	}

	return userID, nil
}

func (s *serverAPI) CreateTask(ctx context.Context, task *tsks_pb.Task) (*tsks_pb.TaskID, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	task.CreatedBy = userID

	if task.GroupID != 0 && task.ActualTime != 0 && task.AssignedTo == 0 {
		return nil, status.Error(codes.InvalidArgument, "forbidden to create completed task in group without task executor")
	}

	task.ID, err = s.repository.SaveTask(ctx, task)
	if err != nil {
		s.log.Error("error while saving task", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if task.ActualTime != 0 {
		err = s.producer.ProduceTaskPredictionData(ctx, models.GetPredictionData(task))
		if err != nil {
			s.log.Error("error while producing task", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	return &tsks_pb.TaskID{ID: task.ID}, nil
}

func (s *serverAPI) GetTask(ctx context.Context, taskID *tsks_pb.TaskID) (*tsks_pb.Task, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}

	task, err := s.repository.GetTaskByID(ctx, taskID.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while getting task:", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if task.GroupID.Int64 == 0 && task.CreatedBy != userID {
		return nil, status.Error(codes.PermissionDenied, "denied")
	}

	return models.ConvertModelToProto(task), nil
}

func (s *serverAPI) GetTaskList(ctx context.Context, req *tsks_pb.TaskListRequest) (*tsks_pb.TaskList, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}

	tasks, err := s.repository.GetTaskList(ctx, userID, req.Offset)
	if err != nil {
		s.log.Error("error while getting task:", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	resp := make([]*tsks_pb.Task, 0, len(tasks))
	for _, task := range tasks {
		resp = append(resp, models.ConvertModelToProto(task))
	}

	return &tsks_pb.TaskList{
		Tasks: resp,
	}, nil
}

func (s *serverAPI) UpdateTask(ctx context.Context, task *tsks_pb.Task) (*emptypb.Empty, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	task.CreatedBy = userID

	// get old task
	taskFromDB, err := s.repository.GetTaskByID(ctx, task.ID)
	if err != nil {
		s.log.Error("error while getting task form db (UpdateTask handler)", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	// validate cases
	if taskFromDB.ActualTime != 0 && task.PlannedTime != 0 {
		return nil, status.Error(codes.FailedPrecondition, "forbidden to change planned time after task completed")
	}
	if taskFromDB.ActualTime != 0 && task.AssignedTo != 0 {
		return nil, status.Error(codes.FailedPrecondition, "forbidden to change user after task completed")
	}
	if taskFromDB.AssignedTo.Int64 == 0 && task.ActualTime != 0 {
		return nil, status.Error(codes.FailedPrecondition, "forbidden to complete task without assigned user")
	}
	if taskFromDB.GroupID.Int64 == 0 && task.AssignedTo != 0 {
		return nil, status.Error(codes.FailedPrecondition, "forbidden to assign task executor outside group")
	}

	if task.GroupID != 0 {
		err = s.repository.UpdateGroupTask(ctx, task)
	} else {
		err = s.repository.UpdateTask(ctx, task)
	}

	if err != nil {
		s.log.Error("error while updating task", sl.Err(err))
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		if errors.Is(err, repository.ErrTaskNotInGroup) {
			return nil, status.Error(codes.InvalidArgument, "task not in group assined_to isn't required")
		}

		return nil, err
	}

	// produce message if task complete, or completed time changed
	if task.ActualTime != 0 {
		err = s.producer.ProduceTaskPredictionData(ctx, models.CombineTasksToPredictionData(
			taskFromDB,
			task,
		))
		if err != nil {
			s.log.Error("error while producing task", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *serverAPI) DeleteTask(ctx context.Context, taskID *tsks_pb.TaskID) (*emptypb.Empty, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}

	// get task
	taskToDelete, err := s.repository.GetTaskByID(ctx, taskID.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &emptypb.Empty{}, nil
		}

		s.log.Error("error while getting task:", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	// delete task
	if taskID.GroupID != 0 {
		// authenticated in gateway
		err = s.repository.DeleteGroupTask(ctx, taskID.ID)
	} else {
		err = s.repository.DeleteUserTask(ctx, userID, taskID.ID)
	}
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while deleting task", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	// produce message to predictions service
	// and save task, in case if produce failed
	err = s.producer.ProduceTaskPredictionDataDelete(ctx, taskID.ID)
	if err != nil {
		s.log.Error("error while produce task data delete message", sl.Err(err))
		_, err := s.repository.SaveTask(ctx, models.ConvertModelToProto(taskToDelete))
		if err != nil {
			s.log.Error("error while saving task", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}

		s.log.Error("error while saving deleted task", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *serverAPI) PredictTask(ctx context.Context, taskID *tsks_pb.TaskID) (*tsks_pb.PredictedTask, error) {
	task, err := s.repository.GetTaskByID(ctx, taskID.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while getting task:", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if task.GroupID.Int64 != 0 && task.AssignedTo.Int64 == 0 {
		return nil, status.Error(codes.FailedPrecondition, "task doesn't have assigned user")
	}
	s.log.Debug("Task form db", slog.Any("task", task))
	predictedTime, err := s.prdtClient.Predict(ctx, task)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal")
	}

	predicted := true
	if predictedTime == 0 {
		predicted = false
	}

	return &tsks_pb.PredictedTask{
		PredictedTime: predictedTime,
		Predicted:     predicted,
		Task:          models.ConvertModelToProto(task),
	}, nil
}

// internal services API

func (s *serverAPI) TaskExists(ctx context.Context, req *tsks_pb.TaskExistsRequest) (*emptypb.Empty, error) {
	if err := s.repository.TaskInGroup(ctx, req.GroupID, req.TaskID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		s.log.Error("error while checking task in group", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *serverAPI) GetGroupTasks(ctx context.Context, groupID *tsks_pb.GroupID) (*tsks_pb.TaskList, error) {
	tasks, err := s.repository.GetGroupTasks(ctx, groupID.ID)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.Internal, "internal")
	}

	resp := make([]*tsks_pb.Task, 0, len(tasks))
	for _, task := range tasks {
		resp = append(resp, models.ConvertModelToProto(task))
	}

	return &tsks_pb.TaskList{
		Tasks: resp,
	}, nil
}

func (s *serverAPI) GetPredictedTasks(ctx context.Context, taskIDs *tsks_pb.TasksIDs) (*tsks_pb.PredictedTaskList, error) {
	tasks, err := s.repository.GetTasks(ctx, taskIDs.IDs)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "tasks not found")
		}

		s.log.Error("error while getting tasks times", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}
	s.log.Debug("tasks", slog.Any("tasks", tasks))

	if len(tasks) != len(taskIDs.IDs) {
		collected := make(map[int64]struct{}, len(tasks))
		for _, t := range tasks {
			collected[t.ID] = struct{}{}
		}

		badIDs := make([]int64, 0, len(taskIDs.IDs)-len(tasks))

		for _, taskID := range taskIDs.IDs {
			if _, ok := collected[taskID]; ok {
				badIDs = append(badIDs, taskID)
			}
		}

		return nil, status.Error(codes.NotFound, fmt.Sprintf("some ids not found %v", badIDs))
	}

	predictedTasks, unpredictedUIDs, err := s.prdtClient.PredictList(ctx, tasks)
	if err != nil {
		if errors.Is(err, predictionsclient.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		s.log.Error("error while predicting task list", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &tsks_pb.PredictedTaskList{
		Tasks:           predictedTasks,
		UnpredictedUIDs: unpredictedUIDs,
	}, nil
}
