package entities

import "github.com/liriquew/control_system/api/internal/models"

type PredictedTask struct {
	Task          models.Task `json:"task"`
	PredictedTime float64     `json:"predicted_time"`
	Predicted     bool        `json:"is_predicted"`
}
