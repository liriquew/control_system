package models

import (
	"encoding/json"
	"io"
)

type Graph struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	GroupID   int64  `json:"group_id" db:"group_id"`
	CreatedBy int64  `json:"created_by" db:"created_by"`
}

type Dependency struct {
	FromNodeID int64 `json:"from_node_id" db:"from_node_id"`
	ToNodeID   int64 `json:"to_node_id" db:"to_node_id"`
}

type Node struct {
	ID                int64   `json:"id" db:"id"`
	GraphID           int64   `json:"graph_id" db:"graph_id"`
	TaskID            int64   `json:"task_id" db:"task_id"`
	AssignedTo        *int64  `json:"assigned_to" db:"assigned_to"`
	DependencyNodeIDs []int64 `json:"dependencies"`
}

func GraphModelFromJson(jsonBody io.ReadCloser) (*Graph, error) {
	var graph Graph
	err := json.NewDecoder(jsonBody).Decode(&graph)

	return &graph, err
}

func NodeModelFromJson(jsonBody io.ReadCloser) (*Node, error) {
	var node Node
	err := json.NewDecoder(jsonBody).Decode(&node)

	return &node, err
}

func DependencyModelFromJson(jsonBody io.ReadCloser) (*Dependency, error) {
	var dependency Dependency
	err := json.NewDecoder(jsonBody).Decode(&dependency)

	return &dependency, err
}
