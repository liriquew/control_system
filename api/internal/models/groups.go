package models

import (
	"encoding/json"
	"io"
	"time"
)

type Group struct {
	ID          int64     `json:"id" db:"id"`
	OwnerID     int64     `json:"owner_id" db:"owner_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type GroupMember struct {
	GroupID int64  `json:"group_id" db:"group_id"`
	UserID  int64  `json:"user_id" db:"user_id"`
	Role    string `json:"role" db:"role"`
}

func GroupModelFromJson(jsonBody io.ReadCloser) (*Group, error) {
	var group Group
	err := json.NewDecoder(jsonBody).Decode(&group)

	return &group, err

}

func GroupMemberModelFromJson(jsonBody io.ReadCloser) (*GroupMember, error) {
	var member GroupMember
	err := json.NewDecoder(jsonBody).Decode(&member)

	return &member, err

}

func (g *Group) Validate() bool {
	return g.Name != "" || g.Description != ""
}
