package models

import (
	"fmt"

	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"
)

type Graph struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	GroupID   int64  `json:"group_id" db:"group_id"`
	CreatedBy int64  `json:"created_by" db:"created_by"`
}

func ConvertGraphToProto(graph *Graph) *grph_pb.Graph {
	return &grph_pb.Graph{
		ID:        graph.ID,
		Name:      graph.Name,
		GroupID:   graph.GroupID,
		CreatedBy: graph.CreatedBy,
	}
}

type Dependency struct {
	FromNodeID int64 `json:"from_node_id" db:"from_node_id"`
	ToNodeID   int64 `json:"to_node_id" db:"to_node_id"`
}

func ConvertDependencyToProto(dep *Dependency) *grph_pb.Dependency {
	return &grph_pb.Dependency{
		FromNodeID: dep.FromNodeID,
		ToNodeID:   dep.ToNodeID,
	}
}

func ConvertDependenciesToProto(deps []*Dependency) []*grph_pb.Dependency {
	res := make([]*grph_pb.Dependency, 0, len(deps))
	for _, dep := range deps {
		res = append(res, ConvertDependencyToProto(dep))
	}
	return res
}

type Node struct {
	ID                     int64   `json:"id" db:"id"`
	GraphID                int64   `json:"graph_id" db:"graph_id"`
	TaskID                 int64   `json:"task_id" db:"task_id"`
	AssignedTo             *int64  `json:"assigned_to" db:"assigned_to"`
	DependencyNodeIDs      []int64 `json:"dependencies"`
	AdditionalDependencies []int64
}

func (n *Node) GetDependencies() []int64 {
	return n.DependencyNodeIDs
}

func (n *Node) GetID() int64 {
	return n.ID
}

func (n *Node) GetAssignedTo() int64 {
	return *n.AssignedTo
}
func (n *Node) GetAdditionalDependencies() []int64 {
	return n.AdditionalDependencies
}

func (n *Node) AddAdditionalDependency(ids ...int64) {
	n.AdditionalDependencies = append(n.AdditionalDependencies, ids...)
}

func ConvertNodeToProto(node *Node) *grph_pb.Node {
	return &grph_pb.Node{
		ID:                node.ID,
		GraphID:           node.GraphID,
		TaskID:            node.TaskID,
		DependencyNodeIDs: node.DependencyNodeIDs,
	}
}

func ConvertNodesToProto(nodes []*Node) []*grph_pb.Node {
	res := make([]*grph_pb.Node, 0, len(nodes))
	for _, node := range nodes {
		fmt.Println(node)
		res = append(res, ConvertNodeToProto(node))
	}
	return res
}

type UserWithTime struct {
	UID  int64
	Time float64
}
