package graphtools

import (
	"fmt"
	"testing"
	"time"

	"github.com/liriquew/graphs_service/internal/entities"
	graph_wrapper "github.com/liriquew/graphs_service/internal/lib/graph_tools/wrapper"
	"github.com/liriquew/graphs_service/internal/models"
)

func ptrInt64(v int64) *int64 { return &v }

func TestFindCriticalPath(t *testing.T) {
	tests := []struct {
		name                 string
		graph                entities.GraphWithTasks
		nodesValueMap        map[int64]float64
		criticalPathNodesIDs [][]int64
	}{
		{
			name: "Simple graph with workers",
			graph: entities.GraphWithTasks{
				Nodes: []*entities.NodeWithTask{
					{
						Node: &models.Node{
							ID:                1,
							DependencyNodeIDs: []int64{2, 3},
							AssignedTo:        ptrInt64(1),
						},
					},
					{
						Node: &models.Node{
							ID:                2,
							DependencyNodeIDs: []int64{4},
							AssignedTo:        ptrInt64(2),
						},
					},
					{
						Node: &models.Node{
							ID:                3,
							DependencyNodeIDs: []int64{5, 6},
							AssignedTo:        ptrInt64(3),
						},
					},
					{
						Node: &models.Node{
							ID:                4,
							DependencyNodeIDs: []int64{6},
							AssignedTo:        ptrInt64(4),
						},
					},
					{
						Node: &models.Node{
							ID:                5,
							DependencyNodeIDs: []int64{6},
							AssignedTo:        ptrInt64(5),
						},
					},
					{
						Node: &models.Node{
							ID:                6,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(6),
						},
					},
				},
			},
			criticalPathNodesIDs: [][]int64{{1, 3, 5, 6}, {1, 3, 6}},
			nodesValueMap: map[int64]float64{
				1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6,
			},
		},
		{
			name: "Single worker with multiple tasks",
			graph: entities.GraphWithTasks{
				Nodes: []*entities.NodeWithTask{
					{
						Node: &models.Node{
							ID:                1,
							DependencyNodeIDs: []int64{2, 3},
							AssignedTo:        ptrInt64(1),
						},
					},
					{
						Node: &models.Node{
							ID:                2,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(2),
						},
					},
					{
						Node: &models.Node{
							ID:                3,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(2),
						},
					},
				},
			},
			criticalPathNodesIDs: [][]int64{{1, 3}, {1, 2, 3}},
			nodesValueMap: map[int64]float64{
				1: 1, 2: 2, 3: 3,
			},
		},
		{
			name: "Additional dependencies test",
			graph: entities.GraphWithTasks{
				Nodes: []*entities.NodeWithTask{
					{
						Node: &models.Node{
							ID:                1,
							DependencyNodeIDs: []int64{2},
							AssignedTo:        ptrInt64(1),
						},
					},
					{
						Node: &models.Node{
							ID:                2,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(2),
						},
					},
					{
						Node: &models.Node{
							ID:                3,
							DependencyNodeIDs: []int64{2},
							AssignedTo:        ptrInt64(3),
						},
					},
				},
			},
			criticalPathNodesIDs: [][]int64{{3, 2}},
			nodesValueMap: map[int64]float64{
				1: 2, 2: 3, 3: 5,
			},
		},
		{
			name: "Multiple critical paths with same duration",
			graph: entities.GraphWithTasks{
				Nodes: []*entities.NodeWithTask{
					{
						Node: &models.Node{
							ID:                1,
							DependencyNodeIDs: []int64{2, 3},
							AssignedTo:        ptrInt64(1),
						},
					},
					{
						Node: &models.Node{
							ID:                2,
							DependencyNodeIDs: []int64{4},
							AssignedTo:        ptrInt64(2),
						},
					},
					{
						Node: &models.Node{
							ID:                3,
							DependencyNodeIDs: []int64{4},
							AssignedTo:        ptrInt64(3),
						},
					},
					{
						Node: &models.Node{
							ID:                4,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(4),
						},
					},
				},
			},
			criticalPathNodesIDs: [][]int64{{1, 2, 4}, {1, 3, 4}},
			nodesValueMap: map[int64]float64{
				1: 2, 2: 3, 3: 3, 4: 2,
			},
		},
		{
			name: "Parallel tasks with dependencies",
			graph: entities.GraphWithTasks{
				Nodes: []*entities.NodeWithTask{
					{
						Node: &models.Node{
							ID:                1,
							DependencyNodeIDs: []int64{2, 3},
							AssignedTo:        ptrInt64(1),
						},
					},
					{
						Node: &models.Node{
							ID:                2,
							DependencyNodeIDs: []int64{5},
							AssignedTo:        ptrInt64(2),
						},
					},
					{
						Node: &models.Node{
							ID:                3,
							DependencyNodeIDs: []int64{5},
							AssignedTo:        ptrInt64(3),
						},
					},
					{
						Node: &models.Node{
							ID:                4,
							DependencyNodeIDs: []int64{5},
							AssignedTo:        ptrInt64(4),
						},
					},
					{
						Node: &models.Node{
							ID:                5,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(5),
						},
					},
				},
			},
			criticalPathNodesIDs: [][]int64{
				{1, 2, 5},
				{1, 3, 5},
			},
			nodesValueMap: map[int64]float64{
				1: 1, 2: 2, 3: 2, 4: 2, 5: 3,
			},
		},
		{
			name: "Big Graph With Worker in few tasks",
			graph: entities.GraphWithTasks{
				Nodes: []*entities.NodeWithTask{
					{
						Node: &models.Node{
							ID:                1,
							DependencyNodeIDs: []int64{2, 3},
							AssignedTo:        ptrInt64(1),
						},
					},
					{
						Node: &models.Node{
							ID:                2,
							DependencyNodeIDs: []int64{4, 5},
							AssignedTo:        ptrInt64(2),
							Weight:            1,
						},
					},
					{
						Node: &models.Node{
							ID:                3,
							DependencyNodeIDs: []int64{5},
							AssignedTo:        ptrInt64(2),
							Weight:            3,
						},
					},
					{
						Node: &models.Node{
							ID:                4,
							DependencyNodeIDs: []int64{5, 6},
							AssignedTo:        ptrInt64(4),
							Weight:            1,
						},
					},
					{
						Node: &models.Node{
							ID:                5,
							DependencyNodeIDs: []int64{6},
							AssignedTo:        ptrInt64(3),
							Weight:            2,
						},
					},
					{
						Node: &models.Node{
							ID:                6,
							DependencyNodeIDs: []int64{},
							AssignedTo:        ptrInt64(5),
							Weight:            3,
						},
					},
				},
			},
			criticalPathNodesIDs: [][]int64{{1, 2, 5, 6}, {1, 3, 5, 6}, {1, 2, 3, 5, 6}},
			nodesValueMap: map[int64]float64{
				1: 1, 2: 4, 3: 3, 4: 1, 5: 2, 6: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan bool)
			go func() {
				result, _ := FindCriticalPath(graph_wrapper.WrapGraphWithTasks(&tt.graph), tt.nodesValueMap)
				if !comparePaths(result, tt.criticalPathNodesIDs) {
					t.Errorf("Expected %v, got %v", tt.criticalPathNodesIDs, result)
				}
				done <- true
			}()

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("Test timed out")
			}
		})
	}
}

func comparePaths(a [][]int64, b [][]int64) bool {
	if len(a) != len(b) {
		return false
	}

	A := make(map[string]int, len(a))
	for _, aa := range a {
		A[fmt.Sprintf("%v", aa)] = 1
	}
	for _, bb := range b {
		bbb := fmt.Sprintf("%v", bb)
		if _, ok := A[bbb]; !ok {
			return false
		}
		A[bbb]--
	}

	for _, v := range A {
		if v != 0 {
			return false
		}
	}

	return true
}
