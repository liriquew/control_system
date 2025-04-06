package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	authclient "github.com/liriquew/tasks_service/internal/grpc/clients/auth_client"
	grphclient "github.com/liriquew/tasks_service/internal/grpc/clients/graphs_client"
	groupsclient "github.com/liriquew/tasks_service/internal/grpc/clients/groups_client"
	grpsclient "github.com/liriquew/tasks_service/internal/grpc/clients/groups_client"
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
	GetTaskList(ctx context.Context, userID, padding int64) ([]*models.Task, error)
	UpdateTask(ctx context.Context, task *tsks_pb.Task) error
	DeleteTask(ctx context.Context, userID, groupID, taskID int64) error

	TaskInAnyGroup(ctx context.Context, taskID int64) (int64, error)
	TaskInGroup(ctx context.Context, groupID, taskID int64) error
	GetTasks(ctx context.Context, taskIDs []int64) ([]*models.Task, error)
	Done(ctx context.Context, taskID int64, time float64) error
}

type tasksProducer interface {
	ProduceTaskPredictionData(context.Context, *models.TaskPredictionData) error
	ProduceTaskPredictionDataDelete(context.Context, int64) error
}

type serverAPI struct {
	tsks_pb.UnimplementedTasksServer
	authClient *authclient.AuthClient
	prdtClient *predictionsclient.PredicionsClient
	grpsClient *grpsclient.GroupsClient
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
	grpsClient *grpsclient.GroupsClient,
	grphClient *grphclient.GraphClient,
) *serverAPI {
	return &serverAPI{
		log:        log,
		repository: taskRepository,
		producer:   producer,
		authClient: authClient,
		prdtClient: prdtClient,
		grpsClient: grpsClient,
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
		return 0, status.Error(codes.Unauthenticated, "missing jwt token")
	}

	AuthParams := md.Get("Authorization")
	if !ok || len(AuthParams) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing jwt token")
	}
	token := AuthParams[0]
	if token == "" {
		return 0, status.Error(codes.Unauthenticated, "missing jwt token")
	}

	token, found := strings.CutPrefix(token, "Bearer ")
	if !found {
		return 0, status.Error(codes.Unauthenticated, "missing jwt token")
	}

	userID, err := s.authClient.Authenticate(ctx, token)
	if err != nil {
		if errors.Is(err, authclient.ErrDeny) {
			return 0, status.Error(codes.Unauthenticated, "denied")
		}
		if errors.Is(err, authclient.ErrMissedJWT) {
			return 0, status.Error(codes.Unauthenticated, "missing jwt")
		}

		s.log.Error("error while authenticate", sl.Err(err))

		return 0, status.Error(codes.Internal, fmt.Errorf("internal error: %w", err).Error())
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

	if task.PlannedTime <= 0 {
		return nil, status.Error(codes.InvalidArgument, "plannedTime must be greated than zero")
	}
	if task.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "empty title")
	}
	if task.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "empty description")
	}
	if task.GroupID != 0 && task.ActualTime != 0 && task.AssignedTo == 0 {
		return nil, status.Error(codes.InvalidArgument, "forbidden to create completed task in group without task executor")
	}

	if task.GroupID != 0 {
		if err := s.grpsClient.CheckAdminPermission(ctx, task.CreatedBy, task.GroupID); err != nil {
			if errors.Is(err, groupsclient.ErrDeny) {
				return nil, status.Error(codes.PermissionDenied, "denied")
			}

			s.log.Error("error checking creator permission in create task handler", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	if task.AssignedTo != 0 {
		if task.GroupID == 0 {
			return nil, status.Error(codes.InvalidArgument, "group_id required")
		}
		if err := s.grpsClient.CheckMemberPermission(ctx, task.AssignedTo, task.GroupID); err != nil {
			if errors.Is(err, groupsclient.ErrDeny) {
				return nil, status.Error(codes.NotFound, "assigned user not in group")
			}

			s.log.Error("error checking member permission in create task handler", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	taskID, err := s.repository.SaveTask(ctx, task)
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

	return &tsks_pb.TaskID{ID: taskID}, nil
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

	if task.GroupID.Int64 != 0 {
		if err := s.grpsClient.CheckMemberPermission(ctx, userID, task.GroupID.Int64); err != nil {
			if errors.Is(err, groupsclient.ErrDeny) {
				return nil, status.Error(codes.PermissionDenied, "denied")
			}

			s.log.Error("error checking creator permission in create task handler", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
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

	tasks, err := s.repository.GetTaskList(ctx, userID, req.Padding)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

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

	// check editor permission
	if err := s.grpsClient.CheckEditorPermission(ctx, userID, taskFromDB.GroupID.Int64); err != nil {
		if errors.Is(err, groupsclient.ErrDeny) {
			return nil, status.Error(codes.PermissionDenied, "user not editor or admin")
		}

		s.log.Error("error checking member permission in update task handler", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	// validate cases
	if task.Title == "" && task.Description == "" && task.PlannedTime == 0.0 && task.ActualTime == 0 && task.AssignedTo == 0 {
		return nil, status.Error(codes.InvalidArgument, "nothing to update")
	}
	if task.PlannedTime < 0 {
		return nil, status.Error(codes.InvalidArgument, "plannedTime must be greated than zero")
	}

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

	// if task must be assigned, check is user in group
	if task.AssignedTo != 0 {
		if err := s.grpsClient.CheckMemberPermission(ctx, task.AssignedTo, taskFromDB.GroupID.Int64); err != nil {
			if errors.Is(err, groupsclient.ErrDeny) {
				return nil, status.Error(codes.NotFound, "assigned user not in group")
			}

			s.log.Error("error checking member permission in create task handler", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	// update task
	err = s.repository.UpdateTask(ctx, task)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		if errors.Is(err, repository.ErrTaskNotInGroup) {
			return nil, status.Error(codes.InvalidArgument, "task not in group assined_to isn't required")
		}

		s.log.Error("error while updating task", sl.Err(err))
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
			return &emptypb.Empty{}, status.Error(codes.OK, "task not found")
		}

		s.log.Error("error while getting task:", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	// check permissions if task in group
	if taskToDelete.GroupID.Valid && taskToDelete.GroupID.Int64 != 0 {
		if err := s.grpsClient.CheckAdminPermission(ctx, userID, taskToDelete.GroupID.Int64); err != nil {
			if errors.Is(err, groupsclient.ErrDeny) {
				return nil, status.Error(codes.PermissionDenied, "denied")
			}

			s.log.Error("error while checking group admin permission", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}

		if nodeID, err := s.grphClient.TaskInNode(ctx, taskID.ID); err != nil {
			if errors.Is(err, grphclient.ErrTaskInNode) {
				return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("task in node %d", nodeID))
			}

			s.log.Error("error while checking is task in some node", sl.Err(err))
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	// delete task
	if err := s.repository.DeleteTask(ctx, userID, taskToDelete.GroupID.Int64, taskID.ID); err != nil {
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

func (s *serverAPI) TaskDone(ctx context.Context, req *tsks_pb.TaskDoneRequest) (*emptypb.Empty, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if req.Time <= 0 {
		return nil, status.Error(codes.InvalidArgument, "time must be greater than zero")
	}

	task, err := s.repository.GetTaskByID(ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while getting task:", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if task.AssignedTo.Int64 != userID && task.CreatedBy != userID {
		return nil, status.Error(codes.PermissionDenied, "only assigned user or creator can complete task")
	}

	if err := s.repository.Done(ctx, req.TaskID, req.Time); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while task done", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	// produce message
	if task.GroupID.Int64 != 0 {
		userID = task.AssignedTo.Int64
	} else {
		userID = task.CreatedBy
	}
	err = s.producer.ProduceTaskPredictionData(ctx, &models.TaskPredictionData{
		ID:          req.TaskID,
		UserID:      userID,
		PlannedTime: task.PlannedTime,
		ActualTime:  req.Time,
	})
	if err != nil {
		s.log.Error("error while producing complete time", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *serverAPI) PredictTask(ctx context.Context, taskID *tsks_pb.TaskID) (*tsks_pb.PredictedTask, error) {
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

	if userID != task.CreatedBy && userID != task.AssignedTo.Int64 {
		return nil, status.Error(codes.PermissionDenied, "")
	}
	if task.GroupID.Int64 != 0 && task.AssignedTo.Int64 == 0 {
		return nil, status.Error(codes.FailedPrecondition, "task doesn't have assigned user")
	}
	s.log.Debug("Task form db", slog.Any("task", task))
	predictedTime, err := s.prdtClient.Predict(ctx, task)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal")
	}

	return &tsks_pb.PredictedTask{
		PredictedTime: predictedTime,
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
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "group tasks not found")
		}

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
