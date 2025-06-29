package graph_wrapper

import (
	"github.com/liriquew/control_system/graphs_service/internal/entities"

	graphtools "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/graph_interface"
)

type WrapperGraphWithTasks struct {
	Nodes []graphtools.Node
}

func (w *WrapperGraphWithTasks) Len() int {
	return len(w.Nodes)
}

func (w *WrapperGraphWithTasks) GetNodes() []graphtools.Node {
	return w.Nodes
}

func WrapGraphWithTasks(g *entities.GraphWithTasks) graphtools.GraphWithNodes {
	n := make([]graphtools.Node, len(g.Nodes))
	for i, node := range g.Nodes {
		n[i] = node
	}
	return &WrapperGraphWithTasks{
		Nodes: n,
	}
}
