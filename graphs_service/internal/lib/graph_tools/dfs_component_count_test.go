package graphtools

import (
	"fmt"
	"testing"

	"github.com/liriquew/graphs_service/internal/entities"
	graph_wrapper "github.com/liriquew/graphs_service/internal/lib/graph_tools/wrapper"
	"github.com/liriquew/graphs_service/internal/models"
)

func TestCountConnectedComponents(t *testing.T) {
	tests := []struct {
		name     string
		graph    *entities.GraphWithNodes
		expected int
		wantErr  bool
	}{
		{
			name: "Single connected component",
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{3}},
					{ID: 3, DependencyNodeIDs: []int64{}},
				},
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "Multiple connected components",
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{}},
					{ID: 3, DependencyNodeIDs: []int64{4}},
					{ID: 4, DependencyNodeIDs: []int64{}},
				},
			},
			expected: 2,
			wantErr:  false,
		},
		{
			name: "Graph with cycle",
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{3}},
					{ID: 3, DependencyNodeIDs: []int64{1}},
				},
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "Graph with unexpected node in dependencies",
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{3}},
					{ID: 3, DependencyNodeIDs: []int64{99}},
				},
			},
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "Empty graph",
			graph:    &entities.GraphWithNodes{Nodes: []*models.Node{}},
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			count, err := CountConnectedComponents(graph_wrapper.WrapGraphWithNodes(tt.graph))
			if (err != nil) != tt.wantErr {
				fmt.Println(err)
				t.Errorf("CountConnectedComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if count != tt.expected {
				t.Errorf("CountConnectedComponents() = %v, want %v", count, tt.expected)
			}
		})
	}
}
