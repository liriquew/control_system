package graphtools

import (
	"testing"

	"github.com/liriquew/control_system/graphs_service/internal/entities"
	graph_wrapper "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/wrapper"
	"github.com/liriquew/control_system/graphs_service/internal/models"
)

func TestHasCycle(t *testing.T) {
	tests := []struct {
		name  string
		graph *entities.GraphWithNodes
		want  bool
	}{
		{
			name: "No cycle",
			// 3 -> 2 -> 1
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{3}},
					{ID: 3, DependencyNodeIDs: []int64{}},
				},
			},
			want: false,
		},
		{
			name: "Cycle exists",
			// 1 -> 3 -> 2 -> 1
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{3}},
					{ID: 3, DependencyNodeIDs: []int64{1}},
				},
			},
			want: true,
		},
		{
			name: "Disconnected graph with cycle",
			// 2 -> 1
			// 3 <-> 4
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{2}},
					{ID: 2, DependencyNodeIDs: []int64{}},
					{ID: 3, DependencyNodeIDs: []int64{4}},
					{ID: 4, DependencyNodeIDs: []int64{3}},
				},
			},
			want: true,
		},
		{
			name: "Big graph with cycle",

			//	---> 1
			//  |   / \
			//  |  2   4
			//  |  | \ |
			//  -- 3 > 5
			//      \ /
			// 		 6
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{3}},
					{ID: 2, DependencyNodeIDs: []int64{1}},
					{ID: 4, DependencyNodeIDs: []int64{1}},
					{ID: 3, DependencyNodeIDs: []int64{2}},
					{ID: 5, DependencyNodeIDs: []int64{2, 3, 4}},
					{ID: 6, DependencyNodeIDs: []int64{3, 5}},
				},
			},
			want: true,
		},
		{
			name: "Big graph without cycle",

			//       1
			//      / \
			//     2   4
			//     | \ |
			//     3 > 5
			//      \ /
			// 		 6
			graph: &entities.GraphWithNodes{
				Nodes: []*models.Node{
					{ID: 1, DependencyNodeIDs: []int64{}},
					{ID: 2, DependencyNodeIDs: []int64{1}},
					{ID: 4, DependencyNodeIDs: []int64{1}},
					{ID: 3, DependencyNodeIDs: []int64{2}},
					{ID: 5, DependencyNodeIDs: []int64{2, 3, 4}},
					{ID: 6, DependencyNodeIDs: []int64{3, 5}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasCycle(graph_wrapper.WrapGraphWithNodes(tt.graph)); got != tt.want {
				t.Errorf("HasCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}
