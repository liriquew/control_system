package heap

import (
	"testing"

	"github.com/liriquew/control_system/internal/models"
)

// lessFunc сравнивает узлы по их ID
func lessByID(a, b *models.Node) bool {
	return a.ID < b.ID
}

func TestNodesWorkerHeap(t *testing.T) {
	// Создаем тестовые узлы
	nodes := []*models.Node{
		{ID: 1, GraphID: 1, TaskID: 2},
		{ID: 2, GraphID: 1, TaskID: 3},
		{ID: 3, GraphID: 1, TaskID: 1},
	}

	// Создаем кучу с функцией сравнения по ID
	heap := NewNodesWorkerHeap(nodes, lessByID)

	// Проверяем, что Pop возвращает элементы в правильном порядке
	expectedOrder := []int64{3, 2, 1}
	for _, expectedID := range expectedOrder {
		node := heap.Pop()
		if node.ID != expectedID {
			t.Errorf("Expected node with ID %d, got %d", expectedID, node.ID)
		}
	}

	// Проверяем, что куча пуста после извлечения всех элементов
	if len(heap.nodes) != 0 {
		t.Errorf("Heap should be empty, but has %d elements", len(heap.nodes))
	}

	// Добавляем новые элементы и проверяем порядок
	heap.Push(&models.Node{ID: 5, GraphID: 1, TaskID: 4})
	heap.Push(&models.Node{ID: 4, GraphID: 1, TaskID: 5})
	heap.Push(&models.Node{ID: 6, GraphID: 1, TaskID: 6})

	expectedOrder = []int64{6, 5, 4}
	for _, expectedID := range expectedOrder {
		node := heap.Pop()
		if node.ID != expectedID {
			t.Errorf("Expected node with ID %d, got %d", expectedID, node.ID)
		}
	}
}

func TestNodesWorkerHeap_PushNode(t *testing.T) {
	// Создаем тестовые узлы
	nodes := []*models.Node{
		{ID: 1, GraphID: 1, TaskID: 2},
		{ID: 2, GraphID: 1, TaskID: 3},
		{ID: 3, GraphID: 1, TaskID: 1},
	}

	// Создаем кучу с функцией сравнения по ID
	heap := NewNodesWorkerHeap(nodes, lessByID)

	heap.Push(&models.Node{ID: 4, GraphID: 1, TaskID: 10})

	// Проверяем, что Pop возвращает элементы в правильном порядке
	expectedOrder := []int64{4, 3, 2, 1}
	for _, expectedID := range expectedOrder {
		node := heap.Pop()
		if node.ID != expectedID {
			t.Errorf("Expected node with ID %d, got %d", expectedID, node.ID)
		}
	}

	// Проверяем, что куча пуста после извлечения всех элементов
	if len(heap.nodes) != 0 {
		t.Errorf("Heap should be empty, but has %d elements", len(heap.nodes))
	}
}

func TestEmptyHeap(t *testing.T) {
	// Создаем пустую кучу
	heap := NewNodesWorkerHeap([]*models.Node{}, lessByID)

	// Попытка извлечь элемент из пустой кучи должна вернуть nil
	node := heap.Pop()
	if node != nil {
		t.Errorf("Expected nil from empty heap, got %v", node)
	}
}

func TestHeapWithSingleElement(t *testing.T) {
	// Создаем кучу с одним элементом
	node := &models.Node{ID: 1, GraphID: 1, TaskID: 1}
	heap := NewNodesWorkerHeap([]*models.Node{node}, lessByID)

	// Извлекаем элемент и проверяем, что он соответствует ожидаемому
	poppedNode := heap.Pop()
	if poppedNode.ID != node.ID {
		t.Errorf("Expected node with ID %d, got %d", node.ID, poppedNode.ID)
	}

	// Проверяем, что куча пуста после извлечения элемента
	if len(heap.nodes) != 0 {
		t.Errorf("Heap should be empty, but has %d elements", len(heap.nodes))
	}
}
