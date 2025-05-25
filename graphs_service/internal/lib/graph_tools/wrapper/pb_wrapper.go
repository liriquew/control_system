package graph_wrapper

import (
	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"
	grph_tools "github.com/liriquew/graphs_service/internal/lib/graph_tools/tools_interface"
)

type WrapperPredictedGraph struct {
	Nodes []grph_tools.Node
}

func (w *WrapperPredictedGraph) Len() int {
	return len(w.Nodes)
}

func (w *WrapperPredictedGraph) GetNodes() []grph_tools.Node {
	return w.Nodes
}

type WrappedNode struct {
	Node *grph_pb.NodeWithTask
}

func (wn *WrappedNode) GetDependencies() []int64 {
	return wn.Node.Node.DependencyNodeIDs
}

func (wn *WrappedNode) GetID() int64 {
	return wn.Node.Node.ID
}

func (wn *WrappedNode) GetWeight() float64 {
	if wn.Node.PredictedTime == 0 {
		return wn.Node.Task.PlannedTime
	}
	return wn.Node.PredictedTime
}

func (wn *WrappedNode) GetAssignedTo() int64 {
	return wn.Node.Task.AssignedTo
}

func (wn *WrappedNode) GetAdditionalDependencies() []int64 {
	return wn.Node.AdditionalDependencies
}

func (wn *WrappedNode) AddAdditionalDependency(ids ...int64) {
	wn.Node.AdditionalDependencies = append(wn.Node.AdditionalDependencies, ids...)
}

func WrapNode(node *grph_pb.NodeWithTask) grph_tools.Node {
	return &WrappedNode{
		Node: node,
	}
}

func WrapPredictedGraph(g *grph_pb.PredictedGraphResponse) grph_tools.GraphWithNodes {
	n := make([]grph_tools.Node, len(g.Nodes))
	for i, node := range g.Nodes {
		n[i] = WrapNode(node)
	}
	return &WrapperGraphWithTasks{
		Nodes: n,
	}
}
