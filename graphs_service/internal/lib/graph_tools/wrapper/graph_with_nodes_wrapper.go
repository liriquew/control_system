package graph_wrapper

import (
	"github.com/liriquew/control_system/graphs_service/internal/entities"
	graphtools "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/graph_interface"
)

type WrapperGraphWithNodes struct {
	Nodes []graphtools.Node
}

func (w *WrapperGraphWithNodes) Len() int {
	return len(w.Nodes)
}

func (w *WrapperGraphWithNodes) GetNodes() []graphtools.Node {
	return w.Nodes
}

func WrapGraphWithNodes(g *entities.GraphWithNodes) graphtools.GraphWithNodes {
	n := make([]graphtools.Node, len(g.Nodes))
	for i, node := range g.Nodes {
		n[i] = node
	}
	return &WrapperGraphWithNodes{
		Nodes: n,
	}
}
