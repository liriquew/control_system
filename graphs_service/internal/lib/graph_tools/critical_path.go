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

type nodeAmount struct {
	min float64
	max float64
	graphinterface.Node
}

func (n *nodeAmount) SetMin(val float64) {
	if n.min > val {
		n.min = val
	}
}

func (n *nodeAmount) SetMax(val float64) {
	if n.max < val {
		n.max = val
	}
}

type less func(graphinterface.Node, graphinterface.Node) bool

const (
	MinTimePriority = iota
	MaxTimePriority

	dummyNodeID int64 = -1
)

var lessFuncMap map[int]less = map[int]less{
	MinTimePriority: func(n1, n2 graphinterface.Node) bool {
		return n1.GetWeight() < n2.GetWeight()
	},
	MaxTimePriority: func(n1, n2 graphinterface.Node) bool {
		return n1.GetWeight() > n2.GetWeight()
	},
}

var (
	ErrCycleInGraph = errors.New("cycle detected")
)

func getLastNodes(adjacencyList map[int64][]int64) []int64 {
	res := []int64{}

	for node, deps := range adjacencyList {
		if len(deps) == 0 {
			res = append(res, node)
		}
	}

	return res
}

type solver struct {
	// мапа хранит состояние каждой вершины (EF, ES, node)
	nodeAmountMap map[int64]*nodeAmount
	// список смежности
	adjacencyList map[int64][]int64
	// инверсированный список смежности, нужен для обратного хода
	adjacencyListInversed map[int64][]int64
	// веса вершин
	nodesValueMap map[int64]float64

	graph graphinterface.GraphWithNodes
}

func newSolver(graph graphinterface.GraphWithNodes, nodesValueMap map[int64]float64) solver {
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
	}

	s.dropState()

	return s
}

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

func (s *solver) createLastNode() {
	// в случае, если есть несколько вершин, которые не имеют исходящих ребер
	// необходимо добавить последнюю вершину, в которой будет отображено время всего графа

	// find last nodes to add dummy node
	lastNodes := getLastNodes(s.adjacencyList)

	// create dummy node
	var dummyNodeAssignedToID int64 = -1
	s.nodeAmountMap[dummyNodeID] = &nodeAmount{
		Node: &entities.NodeWithTask{
			Node: &models.Node{ID: dummyNodeID, AssignedTo: &dummyNodeAssignedToID},
		},
		min: math.MaxFloat64,
		max: math.SmallestNonzeroFloat64,
	}

	if len(lastNodes) == 1 && lastNodes[0] == -1 {
		// already added, this is drop state call,
		// just for init last node
		return
	}

	s.adjacencyList[dummyNodeID] = []int64{}
	s.adjacencyListInversed[dummyNodeID] = lastNodes
	s.nodesValueMap[dummyNodeID] = 0.0

	// connect dummy node to last nodes in graph
	for _, nodeID := range lastNodes {
		node := s.nodeAmountMap[nodeID].Node
		s.adjacencyList[node.GetID()] = append(s.adjacencyList[node.GetID()], dummyNodeID)
	}
}

func (s *solver) forvard() {
	// прямой ход
	// log.Println(s.adjacencyList)
	s.dropState()
	var queue []int64
	doneNodesSet := make(map[int64]struct{}, len(s.nodesValueMap))

	// определение начальных вершин
	for nodeID, inversedNodeDependencies := range s.adjacencyListInversed {
		if len(inversedNodeDependencies) == 0 {
			queue = append(queue, nodeID)
			s.nodeAmountMap[nodeID].min = math.MaxFloat64
			s.nodeAmountMap[nodeID].max = 0
		}
	}

	// fmt.Println("FORWARD")

	for len(queue) > 0 {
		var nextQueue []int64
		// fmt.Println(queue)

		for _, currNodeID := range queue {
			// check is current node ready to done
			readyToDone := true
			for _, prevNodeID := range s.adjacencyListInversed[currNodeID] {
				if _, ok := doneNodesSet[prevNodeID]; !ok {
					readyToDone = false
					break
				}
			}
			if !readyToDone {
				nextQueue = append(nextQueue, currNodeID)
				continue
			}

			doneNodesSet[currNodeID] = struct{}{}

			currNodeAmount := s.nodeAmountMap[currNodeID]
			currNodeTime := currNodeAmount.max + s.nodesValueMap[currNodeID]

			// check next nodes is their previous nodes complete
			// in any case mark them with time which current node achive
			for _, nextNodeID := range s.adjacencyList[currNodeID] {
				allPrevComplete := true
				// fmt.Println(nextNodeID)
				s.nodeAmountMap[nextNodeID].SetMax(currNodeTime)

				for _, prevForNextNodeID := range s.adjacencyListInversed[nextNodeID] {
					_, ok := doneNodesSet[prevForNextNodeID]
					// fmt.Println("\t", prevForNextNodeID, ok)
					allPrevComplete = allPrevComplete && ok
				}

				if allPrevComplete {
					nextQueue = append(nextQueue, nextNodeID)
				}
			}
		}
		// for k, v := range s.nodeAmountMap {
		// 	fmt.Println("\t", k, v.min, v.max)
		// }
		queue = nextQueue
	}
}

func (s *solver) backward() {
	queue := []int64{dummyNodeID}
	doneNodesSet := make(map[int64]struct{}, len(s.nodeAmountMap))
	s.nodeAmountMap[dummyNodeID].SetMin(s.nodeAmountMap[dummyNodeID].max)

	// fmt.Println("BACKWARD")

	for len(queue) != 0 {
		var nextQueue []int64
		// fmt.Println(queue)
		for _, currNodeID := range queue {
			// check is current node ready to done
			readyToDone := true
			for _, prevNodeID := range s.adjacencyList[currNodeID] {
				if _, ok := doneNodesSet[prevNodeID]; !ok {
					readyToDone = false
					break
				}
			}
			if !readyToDone {
				nextQueue = append(nextQueue, currNodeID)
				continue
			}

			doneNodesSet[currNodeID] = struct{}{}

			currNodeTime := s.nodeAmountMap[currNodeID].min
			// check next nodes is their previous nodes complete
			// in any case mark them with time which current node achive
			for _, nextNodeID := range s.adjacencyListInversed[currNodeID] {
				allPrevComplete := true

				s.nodeAmountMap[nextNodeID].SetMin(currNodeTime - s.nodesValueMap[nextNodeID])
				for _, prevForNextNodeID := range s.adjacencyList[nextNodeID] {
					_, ok := doneNodesSet[prevForNextNodeID]
					allPrevComplete = allPrevComplete && ok
				}

				if allPrevComplete {
					nextQueue = append(nextQueue, nextNodeID)
				}
			}
		}
		// for k, v := range s.nodeAmountMap {
		// 	fmt.Println("\t", k, v.min, v.max)
		// }
		queue = nextQueue
	}
}

func (s *solver) correctIntervals() (addedNewDependency bool) {
	type Interval struct {
		start, end float64
		graphinterface.Node
	}
	workersTimeLine := make(map[int64][]Interval, len(s.nodesValueMap))

	for _, node := range s.nodeAmountMap {
		// log.Println(node.GetAssignedTo(), node.GetID(), node.max, node.GetWeight())
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
		sort.Slice(intervals, func(i, j int) bool {
			if intervals[i].start == intervals[j].start {
				return intervals[i].end < intervals[j].end
			}

			return intervals[i].start < intervals[j].start
		})

		for i := range len(intervals) - 1 {
			if intervals[i].end > intervals[i+1].start {
				// log.Println("add", intervals[i], intervals[i+1])
				s.addDependency(
					intervals[i].Node.GetID(),
					intervals[i+1].Node.GetID(),
				)

				addedNewDependency = true
			}
		}
	}

	return
}

func (s *solver) addDependency(fromId, toId int64) {
	s.adjacencyList[fromId] = append(s.adjacencyList[fromId], toId)
	s.adjacencyListInversed[toId] = append(s.adjacencyListInversed[toId], fromId)
}

func (s *solver) collectPaths() [][]int64 {
	// collect critical paths
	var res [][]int64

	// log.Println(s.adjacencyListInversed)

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
			if amount := s.nodeAmountMap[nextNode]; amount.min == amount.max {
				dfs(nextNode, append(curr, nextNode))
			}
		}
	}

	dfs(dummyNodeID, []int64{})

	return res
}

func FindCriticalPath[T graphinterface.GraphWithNodes](graph T, nodesValueMap map[int64]float64) ([][]int64, error) {
	if HasCycle(graph) {
		return nil, ErrCycleInGraph
	}

	solver := newSolver(graph, nodesValueMap)

	solver.forvard()

	if graphChanged := solver.correctIntervals(); graphChanged {
		// if graph corrected (added new dependencies)
		// calculate forward part again
		solver.forvard()
	}

	solver.backward()

	res := solver.collectPaths()

	return res, nil
}
