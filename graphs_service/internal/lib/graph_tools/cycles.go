package graphtools

import graphinterface "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/graph_interface"

// HasCycle - возвращает true, если граф содержит цикл, false в противном случае
// для обнаружения цикла использует топологическую сортировку
func HasCycle(graph graphinterface.GraphWithNodes) (has bool) {
	visited := make(map[int64]struct{}, graph.Len())

	g := make(map[int64][]int64, graph.Len())
	for _, node := range graph.GetNodes() {
		g[node.GetID()] = node.GetDependencies()
	}

	var dfs func(int64) bool
	dfs = func(node int64) bool {
		if len(g[node]) == 0 {
			return false
		}
		if _, ok := visited[node]; ok {
			return true
		}

		visited[node] = struct{}{}

		for _, v := range g[node] {
			if dfs(v) {
				return true
			}
		}

		g[node] = []int64{}
		return false
	}

	for nodeID := range g {
		if dfs(nodeID) {
			return true
		}
	}

	return false
}
