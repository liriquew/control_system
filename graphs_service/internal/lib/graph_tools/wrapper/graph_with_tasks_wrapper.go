package graph_wrapper

import (
	"github.com/liriquew/control_system/graphs_service/internal/entities"

	grph_tools "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/tools_interface"
)

type WrapperGraphWithTasks struct {
	Nodes []grph_tools.Node
}

func (w *WrapperGraphWithTasks) Len() int {
	return len(w.Nodes)
}

func (w *WrapperGraphWithTasks) GetNodes() []grph_tools.Node {
	return w.Nodes
}

func WrapGraphWithTasks(g *entities.GraphWithTasks) grph_tools.GraphWithNodes {
	n := make([]grph_tools.Node, len(g.Nodes))
	for i, node := range g.Nodes {
		n[i] = node
	}
	return &WrapperGraphWithTasks{
		Nodes: n,
	}
}
