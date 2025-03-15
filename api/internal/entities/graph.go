package entities

import (
	"encoding/json"
	"io"
	"time_manage/internal/models"
)

type GraphWithNodes struct {
	GraphInfo models.Graph   `json:"graph"`
	Nodes     []*models.Node `json:"nodes"`
}

func GraphWithNodesFromJSON(jsonBody io.ReadCloser) (*GraphWithNodes, error) {
	var graph GraphWithNodes
	err := json.NewDecoder(jsonBody).Decode(&graph)
	return &graph, err
}

type NodeWithDependency struct {
	Node              *models.Node `json:"node"`
	DependencyNodeIDs []int64      `json:"dependensies"`
}

func NodeWithDependencyFromJSON(jsonBody io.ReadCloser) (*NodeWithDependency, error) {
	var node NodeWithDependency
	err := json.NewDecoder(jsonBody).Decode(&node)
	return &node, err
}
