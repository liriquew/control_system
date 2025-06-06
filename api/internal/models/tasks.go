package models

import (
	"encoding/json"
	"io"
	"time"
)

func TaskModelFromJson(jsonBody io.ReadCloser) (*Task, error) {
	var task Task
	err := json.NewDecoder(jsonBody).Decode(&task)

	return &task, err
}

type Task struct {
	ID          int64     `json:"id,omitempty" db:"id"`
	CreatedBy   int64     `json:"created_by,omitempty" db:"user_id"`
	AssignedTo  int64     `json:"assigned_to,omitempty"`
	GroupID     int64     `json:"group_id,omitempty"`
	Title       string    `json:"title,omitempty" db:"title"`
	Description string    `json:"description,omitempty" db:"description"`
	PlannedTime float64   `json:"planned_time,omitempty" db:"planned_time"`
	Tags        []int32   `json:"tags,omitempty" db:"tags"`
	ActualTime  float64   `json:"actual_time,omitempty" db:"actual_time"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Tag struct {
	ID          int32   `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Probability float64 `json:"probability"`
}

type PredictedTime struct {
	PredictedTime float64 `json:"predicted_time"`
}
