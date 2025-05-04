package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Task struct {
	ID          int64         `db:"id" json:"id,omitempty"`
	CreatedBy   int64         `db:"created_by" json:"user_id,omitempty"`
	AssignedTo  sql.NullInt64 `db:"assigned_to"`
	GroupID     sql.NullInt64 `db:"group_id"`
	Title       string        `db:"title" json:"title,omitempty"`
	Description string        `db:"description" json:"description,omitempty"`
	Tags        pq.Int32Array `db:"tags" json:"tags,omitempty"`
	PlannedTime float64       `db:"planned_time" json:"planned_time,omitempty"`
	ActualTime  float64       `db:"actual_time" json:"actual_time,omitempty"`
	CreatedAt   time.Time     `db:"created_at"`
}

func ConvertModelToProto(task *Task) *tsks_pb.Task {
	return &tsks_pb.Task{
		ID:          task.ID,
		CreatedBy:   task.CreatedBy,
		AssignedTo:  task.AssignedTo.Int64,
		GroupID:     task.GroupID.Int64,
		Title:       task.Title,
		Description: task.Description,
		Tags:        task.Tags,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		CreatedAt:   timestamppb.New(task.CreatedAt),
	}
}

type TaskPredictionData struct {
	OutboxID    int64         `db:"outbox_id"`
	ID          int64         `db:"id" json:"id,omitempty"`
	UserID      int64         `db:"user_id" json:"user_id,omitempty"`
	PlannedTime float64       `db:"planned_time" json:"planned_time,omitempty"`
	ActualTime  float64       `db:"actual_time" json:"actual_time,omitempty"`
	Tags        pq.Int32Array `db:"tags" json:"tags,omitempty"`
}

func ExtractPredictionData(task *Task) *TaskPredictionData {
	res := &TaskPredictionData{
		ID:          task.ID,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		Tags:        task.Tags,
	}

	if task.GroupID.Int64 != 0 {
		res.UserID = task.AssignedTo.Int64
	} else {
		res.UserID = task.CreatedBy
	}

	return res
}

func GetPredictionData(task *tsks_pb.Task) *TaskPredictionData {
	res := &TaskPredictionData{
		ID:          task.ID,
		PlannedTime: task.PlannedTime,
		ActualTime:  task.ActualTime,
		Tags:        task.Tags,
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
