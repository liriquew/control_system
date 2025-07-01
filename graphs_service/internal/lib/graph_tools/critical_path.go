package graphtools

import (
	"errors"
	"math"
	"slices"
	"sort"

	"github.com/liriquew/control_system/graphs_service/internal/entities"
	graphinterface "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/graph_interface"
	"github.com/liriquew/control_system/graphs_service/internal/models"
)

// Хранит состояние текущей вершины
type nodeAmount struct {
	min                 float64 // Раннее начало (ES - early start)
	max                 float64 // Позднее начало (LS - late start)
	graphinterface.Node         // сама вершина
}

// Обновляет ES вершины
func (n *nodeAmount) SetMin(val float64) {
	if n.min > val {
		n.min = val
	}
}

// Обновляет LS вершины
func (n *nodeAmount) SetMax(val float64) {
	if n.max < val {
		n.max = val
	}
}

type less func(graphinterface.Node, graphinterface.Node) bool

const (
	// Константы для выбора приоритета
	MinTimePriority = iota
	MaxTimePriority

	// В случае, если есть несколько вершин, которые не имеют исходящих ребер,
	// то есть несколько финальных вершин
	// необходимо добавить вершину, которая будет аккумулировать состояния предыдущих
	dummyNodeID int64 = -1
)

// Приоритеты
var lessFuncMap map[int]less = map[int]less{
	MinTimePriority: func(n1, n2 graphinterface.Node) bool {
		// вершины с меньшим весом приоритетнее
		return n1.GetWeight() < n2.GetWeight()
	},
	MaxTimePriority: func(n1, n2 graphinterface.Node) bool {
		// вершины с большим весом приоритетнее
		return n1.GetWeight() > n2.GetWeight()
	},
}

var (
	// Будет возвращено, в случае, если граф содержит цикл
	ErrCycleInGraph = errors.New("cycle detected")
)

// Возвращает список индентификаторов вершин, которые не имеют исходящих ребер
func getLastNodes(adjacencyList map[int64][]int64) []int64 {
	res := []int64{}

	for node, deps := range adjacencyList {
		if len(deps) == 0 {
			res = append(res, node)
		}
	}

	return res
}

// solver - хранит состояние алгоритма поиска критических путей
type solver struct {
	// Хранит состояние каждой вершины, ключ - идентификатор вершины
	nodeAmountMap map[int64]*nodeAmount
	// список смежности
	adjacencyList map[int64][]int64
	// инверсированный список смежности, нужен для обратного хода
	adjacencyListInversed map[int64][]int64
	// веса вершин
	nodesValueMap map[int64]float64

	graph graphinterface.GraphWithNodes

	// функция, задающая приоритет задач
	// в случае, если выполнение каких-то задач накладывается
	priorityFunc less
}

// конструктор для структуры, которая хранит состояние алгоритма
func newSolver(graph graphinterface.GraphWithNodes, nodesValueMap map[int64]float64, priority int) solver {
	adjacencyList := make(map[int64][]int64, graph.Len())
	adjacencyListInversed := make(map[int64][]int64, graph.Len())

	// построение списков
	for _, node := range graph.GetNodes() {
		adjacencyList[node.GetID()] = append(node.GetDependencies(), node.GetAdditionalDependencies()...)

		if _, ok := adjacencyListInversed[node.GetID()]; !ok {
			adjacencyListInversed[node.GetID()] = []int64{}
		}
		for _, depNode := range adjacencyList[node.GetID()] {
			adjacencyListInversed[depNode] = append(adjacencyListInversed[depNode], node.GetID())
		}
	}

	s := solver{
		adjacencyList:         adjacencyList,
		adjacencyListInversed: adjacencyListInversed,
		nodesValueMap:         nodesValueMap,
		graph:                 graph,
		priorityFunc:          lessFuncMap[priority],
	}

	s.dropState()

	return s
}

// dropState - заполняет nodeAmountMap значениями по умолчанию
// и добавляет последнюю вершину
func (s *solver) dropState() {
	s.nodeAmountMap = make(map[int64]*nodeAmount, s.graph.Len())
	for _, node := range s.graph.GetNodes() {
		s.nodeAmountMap[node.GetID()] = &nodeAmount{
			Node: node,
			min:  math.MaxFloat64,
			max:  math.SmallestNonzeroFloat64,
		}
	}

	s.createLastNode()
}

// В случае, если есть несколько вершин, которые не имеют исходящих ребер
// необходимо добавить последнюю вершину, в которой будет отображено время всего графа
func (s *solver) createLastNode() {
	// находим список последних вершин
	lastNodes := getLastNodes(s.adjacencyList)

	// создаем вершину - аккумулятор
	var dummyNodeAssignedToID int64 = -1
	s.nodeAmountMap[dummyNodeID] = &nodeAmount{
		Node: &entities.NodeWithTask{
			Node: &models.Node{ID: dummyNodeID, AssignedTo: &dummyNodeAssignedToID},
		},
		min: math.MaxFloat64,
		max: math.SmallestNonzeroFloat64,
	}

	if len(lastNodes) == 1 && lastNodes[0] == -1 {
		// последняя вершина уже добавлена,
		// этот вызов связан с тем, что структура графа была изменена
		// вызов для того, чтобы добавить последнюю вершину
		return
	}

	// добавляем последнюю вершину
	s.adjacencyList[dummyNodeID] = []int64{}
	s.adjacencyListInversed[dummyNodeID] = lastNodes
	s.nodesValueMap[dummyNodeID] = 0.0

	// добавляем ребра из последних вершин к добавляемой
	for _, nodeID := range lastNodes {
		node := s.nodeAmountMap[nodeID].Node
		s.adjacencyList[node.GetID()] = append(s.adjacencyList[node.GetID()], dummyNodeID)
	}
}

// forward - выполняет прямой ход алгоритма
func (s *solver) forvard() {
	// сбрасываем состояние, добавляем последнюю вершниу
	s.dropState()

	// очередь рассматриваемых вершин
	var queue []int64
	// множество вершин, которые выполнены
	doneNodesSet := make(map[int64]struct{}, len(s.nodesValueMap))

	// Определение начальных вершин
	for nodeID, inversedNodeDependencies := range s.adjacencyListInversed {
		if len(inversedNodeDependencies) == 0 {
			queue = append(queue, nodeID)
			s.nodeAmountMap[nodeID].min = math.MaxFloat64
			s.nodeAmountMap[nodeID].max = 0
		}
	}

	// использоваием алгоритма обхода графа в ширину,
	// обходим все вершины, при этом обновляя ES и EF рассматриваемых вершин
	for len(queue) > 0 {
		var nextQueue []int64

		for _, currNodeID := range queue {
			// проверка, готова ли текущая вершина к выполнению
			// для этого все предыдущие вершины должны быть выполнены
			readyToDone := true
			for _, prevNodeID := range s.adjacencyListInversed[currNodeID] {
				if _, ok := doneNodesSet[prevNodeID]; !ok {
					readyToDone = false
					break
				}
			}

			// если вершина не готова к выполнению
			// рассмотрим ее на следующей интерации
			if !readyToDone {
				nextQueue = append(nextQueue, currNodeID)
				continue
			}

			// текущая вершина выполненена (так как все предыдущие вершины выполнены)
			doneNodesSet[currNodeID] = struct{}{}

			// обновляем состояние текущей вершины
			currNodeAmount := s.nodeAmountMap[currNodeID]
			currNodeTime := currNodeAmount.max + s.nodesValueMap[currNodeID]

			// обходим вершины, в которые можно перейти из текущей вершины
			for _, nextNodeID := range s.adjacencyList[currNodeID] {
				// обновляем состоние вершины
				s.nodeAmountMap[nextNodeID].SetMax(currNodeTime)
				// добавляем в очередь
				nextQueue = append(nextQueue, nextNodeID)
			}
		}
		// обновляем очередь для следующей итерации
		queue = nextQueue
	}
}

// forward - выполняет прямой ход алгоритма
func (s *solver) backward() {
	// очередь рассматриваемых вершин,
	// изначально в очереди аккумулирующая вершина
	queue := []int64{dummyNodeID}

	// обновляем состояние акуккумулирующей вершины,
	// ее LS должен быть равен ES
	s.nodeAmountMap[dummyNodeID].SetMin(s.nodeAmountMap[dummyNodeID].max)

	// использоваием алгоритма обхода графа в ширину,
	// обходим все вершины, при этом обновляя ES и EF рассматриваемых вершин
	for len(queue) != 0 {
		// очередь для следующей итерации
		var nextQueue []int64
		for _, currNodeID := range queue {
			// получаем время текущей вершины (ee LS)
			currNodeTime := s.nodeAmountMap[currNodeID].min

			for _, nextNodeID := range s.adjacencyListInversed[currNodeID] {
				// для каждой предшествующей вершины
				// LS = min(LS текущей - время выполнения текущей)
				s.nodeAmountMap[nextNodeID].SetMin(currNodeTime - s.nodesValueMap[nextNodeID])

				// добавляем вершину в очередь
				nextQueue = append(nextQueue, nextNodeID)
			}
		}
		// обновляем очередь для следующей итерации
		queue = nextQueue
	}
}

// correctIntervals - выполняет корректировку в графе, добавляя ребра (зависимости)
// в случае, если для какого-то исполнителя выявились накладывающиеся задачи
func (s *solver) correctIntervals() (addedNewDependency bool) {
	// Interval - хранит информацию о временном интервале, в котором исполнитель выполнял задачу
	type Interval struct {
		// начало и конец интервала
		start, end float64
		// вершина обозначающая задачу, которая выполняется в этом интервале
		graphinterface.Node
	}

	workersTimeLine := make(map[int64][]Interval, len(s.nodesValueMap))

	// группируем интервалы по исполнителям
	for _, node := range s.nodeAmountMap {
		workersTimeLine[node.GetAssignedTo()] = append(
			workersTimeLine[node.GetAssignedTo()],
			Interval{
				start: node.max,
				end:   node.max + node.GetWeight(),
				Node:  node,
			},
		)
	}

	for _, intervals := range workersTimeLine {
		// сортируем интервалы исполнителя
		// по возрастанию времени начала интервала
		sort.Slice(intervals, func(i, j int) bool {
			// в случае, если начало интервалов совпадают,
			// меньше тот, который раньше заканчивается
			if intervals[i].start == intervals[j].start {
				return intervals[i].end < intervals[j].end
			}

			return intervals[i].start < intervals[j].start
		})

		for i := range len(intervals) - 1 {
			// в случае если конец текщуего интервала превосходит начало следующего
			// добавляем в зависимости (i+1)-й задачи i-ю
			if intervals[i].end > intervals[i+1].start {
				s.addDependency(
					intervals[i].Node.GetID(),
					intervals[i+1].Node.GetID(),
				)

				// отмечаем, что структура графа была изменена
				addedNewDependency = true
			}
		}
	}

	return
}

// addDependency - добавляет зависимость
// toId будет зависеть от выполнения вершины с индетификатором fromId
func (s *solver) addDependency(fromId, toId int64) {
	s.adjacencyList[fromId] = append(s.adjacencyList[fromId], toId)
	s.adjacencyListInversed[toId] = append(s.adjacencyListInversed[toId], fromId)
}

// collectPaths - собирает критические пути графа
func (s *solver) collectPaths() [][]int64 {
	var res [][]int64

	// рекурсивно, с помощью алгоритма поиска в грубину обходим граф
	// и собираем вершины у которых ES == EF
	var dfs func(int64, []int64)
	dfs = func(node int64, curr []int64) {
		if amount := s.nodeAmountMap[node]; amount.max == 0 && amount.min == amount.max {
			path := make([]int64, len(curr))
			copy(path, curr)
			slices.Reverse(path)
			res = append(res, path)
			return
		}

		for _, nextNode := range s.adjacencyListInversed[node] {
			// рассматриваем только те пути, которые содержат вершины ES == EF
			// прочие вершины не формируют критический путь
			if amount := s.nodeAmountMap[nextNode]; amount.min == amount.max {
				dfs(nextNode, append(curr, nextNode))
			}
		}
	}

	dfs(dummyNodeID, []int64{})

	return res
}

// FindCriticalPath - находит критические пути графа
// graph - граф
// nodesValueMap - определяет веса вершин
func FindCriticalPath[T graphinterface.GraphWithNodes](graph T, nodesValueMap map[int64]float64, priority int) ([][]int64, error) {
	// в случае, если в графе есть цикл, вернем ошибку
	if HasCycle(graph) {
		return nil, ErrCycleInGraph
	}

	solver := newSolver(graph, nodesValueMap, priority)

	// выполняем рассчет прямого хода
	solver.forvard()

	// проверяем наличие накладывающихся задач
	if graphChanged := solver.correctIntervals(); graphChanged {
		// если граф был изменен (были накладывающиеся задачи)
		// необходимо вычислить прямой ход повторно
		solver.forvard()
	}

	// рассчет обратного хода
	solver.backward()

	// сбор критических путей
	res := solver.collectPaths()

	return res, nil
}
