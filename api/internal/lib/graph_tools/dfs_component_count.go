package graphtools

import (
	"errors"
	"fmt"

	"github.com/liriquew/control_system/internal/entities"
)

var (
	ErrUnexpectedNodeInDeps = errors.New("got unexpected node in dependencies")
)

func CountConnectedComponents(graph entities.GraphWithNodes) (int, error) {
	visited := make(map[int64]struct{}, len(graph.Nodes))

	g := make(map[int64][]int64, len(graph.Nodes))
	for _, node := range graph.Nodes {
		g[node.ID] = node.DependencyNodeIDs
	}

	var err error
	var dfs func(int64)
	dfs = func(node int64) {
		if _, ok := g[node]; !ok {
			err = fmt.Errorf(
				"error while counting components, error node: %d error: %w",
				node,
				ErrUnexpectedNodeInDeps,
			)
			return
		}

		visited[node] = struct{}{}

		for _, nodeID := range g[node] {
			if _, ok := visited[nodeID]; !ok {
				dfs(nodeID)
			}
		}
	}

	count := 0
	for node := range g {
		if _, ok := visited[node]; !ok {
			count++
			dfs(node)
			if err != nil {
				return 0, err
			}
		}
	}

	return count, nil
}
