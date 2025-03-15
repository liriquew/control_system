package models

import (
	"encoding/json"
	"io"
	"time"
)

type User struct {
	ID       int64  `json:"uid" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type Task struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	PlannedTime float64   `json:"planned_time" db:"planned_time"`
	ActualTime  *float64  `json:"actual_time" db:"actual_time"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

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

func UserModelFromJson(jsonBody io.ReadCloser) (*User, error) {
	var user User
	err := json.NewDecoder(jsonBody).Decode(&user)

	return &user, err
}

func TaskModelFromJson(jsonBody io.ReadCloser) (*Task, error) {
	var task Task
	err := json.NewDecoder(jsonBody).Decode(&task)

	return &task, err
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
