package models

import (
	"database/sql"
	"time"

	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Task struct {
	ID          int64         `db:"id" json:"ID,omitempty"`
	CreatedBy   int64         `db:"created_by" json:"CreatedBy,omitempty"`
	AssignedTo  sql.NullInt64 `db:"assigned_to"`
	GroupID     sql.NullInt64 `db:"group_id" json:"GroupID,omitempty"`
	Title       string        `db:"title" json:"Title,omitempty"`
	Description string        `db:"description" json:"Description,omitempty"`
	PlannedTime float64       `db:"planned_time" json:"PlannedTime,omitempty"`
	ActualTime  float64       `db:"actual_time" json:"ActualTime,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"CreatedAt,omitempty"`
}

func ConvertModelToProto(task *Task) *tsks_pb.Task {
	return &tsks_pb.Task{
		ID:          task.ID,
		CreatedBy:   task.CreatedBy,
		AssignedTo:  task.AssignedTo.Int64,
		GroupID:     task.GroupID.Int64,
		Title:       task.Title,
		Description: task.Description,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		CreatedAt:   timestamppb.New(task.CreatedAt),
	}
}

type TaskPredictionData struct {
	ID          int64   `json:"ID,omitempty"`
	UserID      int64   `json:"UserID,omitempty"`
	PlannedTime float64 `json:"PlannedTime,omitempty"`
	ActualTime  float64 `json:"ActualTime,omitempty"`
}

func GetPredictionData(task *tsks_pb.Task) *TaskPredictionData {
	res := &TaskPredictionData{
		ID:          task.ID,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
	}

	if task.GroupID != 0 {
		res.UserID = task.AssignedTo
	} else {
		res.UserID = task.CreatedBy
	}

	return res
}
func CombineTasksToPredictionData(task1 *Task, task2 *tsks_pb.Task) *TaskPredictionData {
	res := &TaskPredictionData{
		ID: task1.ID,
	}

	if task1.GroupID.Int64 != 0 {
		// group task
		res.UserID = task1.AssignedTo.Int64
	} else {
		// self task
		res.UserID = task2.CreatedBy
	}

	if task2.ActualTime != 0 {
		res.ActualTime = task2.ActualTime
	} else {
		res.ActualTime = task1.ActualTime
	}

	if task1.PlannedTime != 0 {
		res.PlannedTime = task1.PlannedTime
	} else {
		res.PlannedTime = task2.PlannedTime
	}

	return res
}
