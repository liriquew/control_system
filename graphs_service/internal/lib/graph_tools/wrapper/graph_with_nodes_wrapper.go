package graph_wrapper

import (
	"github.com/liriquew/graphs_service/internal/entities"

	grph_tools "github.com/liriquew/graphs_service/internal/lib/graph_tools/tools_interface"
)

type WrapperGraphWithNodes struct {
	Nodes []grph_tools.Node
}

func (w *WrapperGraphWithNodes) Len() int {
	return len(w.Nodes)
}

func (w *WrapperGraphWithNodes) GetNodes() []grph_tools.Node {
	return w.Nodes
}

func WrapGraphWithNodes(g *entities.GraphWithNodes) grph_tools.GraphWithNodes {
	n := make([]grph_tools.Node, len(g.Nodes))
	for i, node := range g.Nodes {
		n[i] = node
	}
	return &WrapperGraphWithNodes{
		Nodes: n,
	}
}
