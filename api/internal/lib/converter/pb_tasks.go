package converter

import (
	"github.com/liriquew/control_system/internal/entities"
	"github.com/liriquew/control_system/internal/models"
	prdt_pb "github.com/liriquew/control_system/services_protos/predictions_service"
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
		Tags:        task.Tags,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		CreatedAt:   timestamppb.New(task.CreatedAt),
	}
}

func ConvertTaskToModel(task *tsks_pb.Task) *models.Task {
	return &models.Task{
		ID:          task.ID,
		CreatedBy:   task.CreatedBy,
		AssignedTo:  task.AssignedTo,
		GroupID:     task.GroupID,
		Title:       task.Title,
		Description: task.Description,
		Tags:        task.Tags,
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

func ConvertTagToModel(tag *prdt_pb.Tag) *models.Tag {
	return &models.Tag{
		ID:          tag.Id,
		Name:        tag.Name,
		Probability: tag.Probability,
	}
}

func ConvertTagsToModel(tags []*prdt_pb.Tag) []*models.Tag {
	res := make([]*models.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, ConvertTagToModel(tag))
	}
	return res
}
