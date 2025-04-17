package converter

import (
	"fmt"

	"github.com/liriquew/control_system/internal/entities"
	"github.com/liriquew/control_system/internal/models"
	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertTaskToProto(task *models.Task) *tsks_pb.Task {
	return &tsks_pb.Task{
		ID:          task.ID,
		CreatedBy:   task.CreatedBy,
		AssignedTo:  task.AssignedTo,
		GroupID:     task.GroupID,
		Title:       task.Title,
		Description: task.Description,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		CreatedAt:   timestamppb.New(task.CreatedAt),
	}
}

func ConvertTaskToModel(task *tsks_pb.Task) *models.Task {
	fmt.Println(task.GroupID)
	return &models.Task{
		ID:          task.ID,
		CreatedBy:   task.CreatedBy,
		AssignedTo:  task.AssignedTo,
		GroupID:     task.GroupID,
		Title:       task.Title,
		Description: task.Description,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		CreatedAt:   task.CreatedAt.AsTime(),
	}
}

func ConvertTasksToModel(tasks []*tsks_pb.Task) []*models.Task {
	res := make([]*models.Task, 0, len(tasks))
	for _, task := range tasks {
		res = append(res, ConvertTaskToModel(task))
	}
	return res
}

func ConvertPredictedTaskToModel(predictedTask *tsks_pb.PredictedTask) *entities.PredictedTask {
	return &entities.PredictedTask{
		Task:          *ConvertTaskToModel(predictedTask.Task),
		PredictedTime: predictedTask.PredictedTime,
		Predicted:     predictedTask.Predicted,
	}
}
