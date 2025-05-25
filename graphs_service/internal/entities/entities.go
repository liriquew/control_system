package entities

import (
	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"github.com/liriquew/graphs_service/internal/models"
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

type NodeWithTask struct {
	Node                   *models.Node  `json:"node"`
	Task                   *tsks_pb.Task `json:"task"`
	AdditionalDependencies []int64       `json:"additional_priority_deps"`
	PredictedTime          float64       `json:"predicted_time"`
}

func (nt *NodeWithTask) GetDependencies() []int64 {
	return nt.Node.DependencyNodeIDs
}

func (nt *NodeWithTask) GetID() int64 {
	return nt.Node.ID
}

func (nt *NodeWithTask) GetWeight() float64 {
	return nt.Node.Weight
}

func (nt *NodeWithTask) GetAssignedTo() int64 {
	return *nt.Node.AssignedTo
}

func (nt *NodeWithTask) GetAdditionalDependencies() []int64 {
	return nt.AdditionalDependencies
}

func (nt *NodeWithTask) AddAdditionalDependency(ids ...int64) {
	nt.AdditionalDependencies = append(nt.AdditionalDependencies, ids...)
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
	Graph *GraphWithTasks
	Paths [][]int64
}

type NodeWithDependency struct {
	Node              *models.Node `json:"node"`
	DependencyNodeIDs []int64      `json:"dependensies"`
}
