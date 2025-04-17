package entities

import (
	"encoding/json"
	"io"

	"github.com/liriquew/control_system/internal/models"
)

type GraphWithNodes struct {
	GraphInfo *models.Graph  `json:"graph"`
	Nodes     []*models.Node `json:"nodes"`
}

func (g *GraphWithNodes) Len() int {
	return len(g.Nodes)
}

func (g *GraphWithNodes) GetNodes() []*models.Node {
	return g.Nodes
}

func GraphWithNodesFromJSON(jsonBody io.ReadCloser) (*GraphWithNodes, error) {
	var graph GraphWithNodes
	err := json.NewDecoder(jsonBody).Decode(&graph)
	return &graph, err
}

type NodeWithTask struct {
	Node                   *models.Node `json:"node"`
	Task                   *models.Task `json:"task"`
	AdditionalDependencies []int64      `json:"additional_priority_deps"`
	PredictedTime          float64      `json:"predicted_time"`
}

type PredictedNodes struct {
	Nodes           []*NodeWithTask
	UnpredictedUIDs []int64
}

type GraphWithTasks struct {
	GraphInfo       models.Graph    `json:"graph"`
	Nodes           []*NodeWithTask `json:"nodes"`
	UnpredictedUIDs []int64         `json:"unpredictedUIDs"`
}

func (g *GraphWithTasks) Len() int {
	return len(g.Nodes)
}

func (g *GraphWithTasks) GetNodes() []*models.Node {
	var nodes []*models.Node
	for _, nodeWithTask := range g.Nodes {
		nodes = append(nodes, nodeWithTask.Node)
	}
	return nodes
}

type PredictedGraph struct {
	Graph           *models.Graph
	Nodes           []*NodeWithTask
	UnpredictedUIDs []int64
	Paths           [][]int64
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
