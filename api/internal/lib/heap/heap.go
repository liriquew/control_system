package heap

import (
	"github.com/liriquew/control_system/internal/models"
)

// max heap для models.Node
type NodesWorkerHeap struct {
	nodes      []*models.Node
	addedNodes map[int64]struct{}
	lessFunc   func(*models.Node, *models.Node) bool
}

// NewNodesWorkerHeap возвращает кучу для *models.Node.
// Параметр less — функция сравнения, которая возвращает true, если n1 имеет меньший приоритет, чем n2 (для max heap).
// Если less возвращает true, когда n1 имеет больший приоритет, чем n2, то это будет max heap.
// По умолчанию реализация поддерживает min heap.
func NewNodesWorkerHeap(initNodes []*models.Node, less func(*models.Node, *models.Node) bool) *NodesWorkerHeap {
	h := NodesWorkerHeap{
		nodes:      initNodes,
		addedNodes: make(map[int64]struct{}),
		lessFunc:   less,
	}

	h.init()

	return &h
}

func (nh *NodesWorkerHeap) less(i, j int) bool {
	return nh.lessFunc(nh.nodes[i], nh.nodes[j])
}

func (nh *NodesWorkerHeap) swap(i, j int) {
	nh.nodes[i], nh.nodes[j] = nh.nodes[j], nh.nodes[i]
}

func (nh *NodesWorkerHeap) Len() int {
	return len(nh.nodes)
}

func (nh *NodesWorkerHeap) Push(node *models.Node) {
	nh.addedNodes[node.ID] = struct{}{}
	nh.nodes = append(nh.nodes, node)
	nh.up(len(nh.nodes) - 1)
}

func (nh *NodesWorkerHeap) Pop() *models.Node {
	n := len(nh.nodes) - 1
	if n == -1 {
		return nil
	}
	nh.swap(0, n)
	nh.down(0, n)
	node := nh.nodes[n]
	nh.nodes = nh.nodes[:n]
	return node
}

func (nh *NodesWorkerHeap) Top() *models.Node {
	if len(nh.nodes) == 0 {
		return nil
	}
	return nh.nodes[len(nh.nodes)-1]
}

func (nh *NodesWorkerHeap) init() {
	n := len(nh.nodes)
	for i := n/2 - 1; i >= 0; i-- {
		nh.down(i, n)
	}
}

func (nh *NodesWorkerHeap) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n {
			break
		}
		j := j1
		if j2 := j1 + 1; j2 < n && nh.less(j1, j2) {
			j = j2
		}
		if !nh.less(i, j) {
			break
		}
		nh.swap(i, j)
		i = j
	}
}

func (nh *NodesWorkerHeap) up(j int) {
	for {
		i := (j - 1) / 2
		if i == j || !nh.less(i, j) {
			break
		}
		nh.swap(i, j)
		j = i
	}
}
