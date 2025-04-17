package entities

import "github.com/liriquew/control_system/internal/models"

type GroupWithTasks struct {
	Group *models.Group  `json:"group"`
	Tasks []*models.Task `json:"tasks"`
}

type GroupMemberWithDetails struct {
	Member  *models.GroupMember  `json:"member"`
	Details *models.UsersDetails `json:"details"`
}
