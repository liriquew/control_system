package graphtools

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"

	"github.com/liriquew/graphs_service/internal/entities"
	graph_tools_interface "github.com/liriquew/graphs_service/internal/lib/graph_tools/tools_interface"
	"github.com/liriquew/graphs_service/internal/models"
)

const dummyNodeID int64 = -1

var (
	ErrCycleInGraph = errors.New("cycle detected")
)

type nodeAmount struct {
	min  float64
	max  float64
	node graph_tools_interface.Node
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

type less func(graph_tools_interface.Node, graph_tools_interface.Node) bool

var lessFuncMap map[int]less

func init() {
	lessFuncMap = make(map[int]less)
	lessFuncMap[1] = func(n1, n2 graph_tools_interface.Node) bool {
		return n1.GetID() < n2.GetID()
	}
}

func addPriorityDependencies(graph graph_tools_interface.GraphWithNodes, nodesValueMap map[int64]float64, lf less) {
	workerTasks := make(map[int64][]graph_tools_interface.Node)

	// Группируем задачи по исполнителям
	for _, node := range graph.GetNodes() {
		if assignedTo := node.GetAssignedTo(); assignedTo != 0 {
			workerTasks[assignedTo] = append(workerTasks[assignedTo], node)
		}
	}

	// Сортируем задачи по убыванию длительности и добавляем зависимости
	for _, tasks := range workerTasks {
		// Сортировка по приоритету (здесь — по длительности)
		sort.SliceStable(tasks, func(i, j int) bool {
			return lf(tasks[i], tasks[i])
		})

		// Добавляем рёбра между последовательными задачами
		for i := 0; i < len(tasks)-1; i++ {
			current := tasks[i]
			next := tasks[i+1]
			current.AddAdditionalDependency(next.GetID())
		}
	}
}

func getLastNodes(adjacencyList map[int64][]int64) []int64 {
	visited := make(map[int64]struct{}, len(adjacencyList))
	res := []int64{}

	var dfs func(int64)
	dfs = func(node int64) {
		if len(adjacencyList[node]) == 0 {
			res = append(res, node)
		}

		for _, v := range adjacencyList[node] {
			if _, ok := visited[node]; ok {
				dfs(v)
			}
		}
	}

	for node := range adjacencyList {
		if _, ok := visited[node]; !ok {
			dfs(node)
		}
	}

	return res
}

func FindCriticalPath(graph graph_tools_interface.GraphWithNodes, nodesValueMap map[int64]float64) ([][]int64, error) {
	if HasCycle(graph) {
		return nil, ErrCycleInGraph
	}

	addPriorityDependencies(graph, nodesValueMap, lessFuncMap[1])

	nodeAmountMap := make(map[int64]*nodeAmount, graph.Len())
	// список смежности
	adjacencyList := make(map[int64][]int64, graph.Len())
	// обратный список смежности (для обратного хода)
	adjacencyListInversed := make(map[int64][]int64, graph.Len())

	// построение списков
	for _, node := range graph.GetNodes() {
		adjacencyList[node.GetID()] = append(node.GetDependencies(), node.GetAdditionalDependencies()...)

		if _, ok := adjacencyListInversed[node.GetID()]; !ok {
			adjacencyListInversed[node.GetID()] = []int64{}
		}
		for _, depNode := range adjacencyList[node.GetID()] {
			if _, ok := adjacencyListInversed[depNode]; !ok {
				adjacencyListInversed[depNode] = []int64{}
			}

			adjacencyListInversed[depNode] = append(adjacencyListInversed[depNode], node.GetID())
		}

		nodeAmountMap[node.GetID()] = &nodeAmount{
			node: node,
			min:  math.MaxFloat64,
			max:  math.SmallestNonzeroFloat64,
		}
	}

	// find last nodes to add dummy node
	lastNodes := getLastNodes(adjacencyList)

	// create dummy node
	var dummyNodeAssignedToID int64 = -1
	lastNode := &entities.NodeWithTask{
		Node: &models.Node{ID: dummyNodeID, AssignedTo: &dummyNodeAssignedToID},
	}
	nodeAmountMap[dummyNodeID] = &nodeAmount{
		node: lastNode,
		min:  math.MaxFloat64,
		max:  math.SmallestNonzeroFloat64,
	}
	adjacencyList[dummyNodeID] = []int64{}
	adjacencyListInversed[dummyNodeID] = lastNodes
	nodesValueMap[dummyNodeID] = 0.0

	// connect dummy node to last nodes in graph
	for _, nodeID := range lastNodes {
		node := nodeAmountMap[nodeID].node
		adjacencyList[node.GetID()] = append(adjacencyList[node.GetID()], dummyNodeID)
	}

	// straight part
	var queue []int64
	doneNodesSet := make(map[int64]struct{}, graph.Len())

	for nodeID, inversedNodeDependencies := range adjacencyListInversed {
		if len(inversedNodeDependencies) == 0 {
			queue = append(queue, nodeID)
			nodeAmountMap[nodeID].min = math.MaxFloat64
			nodeAmountMap[nodeID].max = 0
		}
	}

	for len(queue) > 0 {
		var nextQueue []int64
		fmt.Println(queue)

		for _, currNodeID := range queue {
			// check is current node ready to done
			readyToDone := true
			for _, prevNodeID := range adjacencyListInversed[currNodeID] {
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

			currNodeAmount := nodeAmountMap[currNodeID]
			currNodeTime := currNodeAmount.max + nodesValueMap[currNodeID]
			// check next nodes is their previous nodes complete
			// in any case mark them with time which current node achive
			for _, nextNodeID := range adjacencyList[currNodeID] {
				allPrevComplete := true
				nodeAmountMap[nextNodeID].SetMax(currNodeTime)

				for _, prevForNextNodeID := range adjacencyListInversed[nextNodeID] {
					_, ok := doneNodesSet[prevForNextNodeID]
					allPrevComplete = allPrevComplete && ok
				}

				if allPrevComplete {
					nextQueue = append(nextQueue, nextNodeID)
				}
			}
		}
		for k, v := range nodeAmountMap {
			fmt.Println("\t", k, v.min, v.max)
		}
		queue = nextQueue
	}

	// backward part
	queue = []int64{dummyNodeID}
	doneNodesSet = make(map[int64]struct{}, graph.Len())
	nodeAmountMap[dummyNodeID].SetMin(nodeAmountMap[dummyNodeID].max)

	fmt.Println("BACKWARD")

	for len(queue) != 0 {
		var nextQueue []int64
		fmt.Println(queue)
		for _, currNodeID := range queue {
			// check is current node ready to done
			readyToDone := true
			for _, prevNodeID := range adjacencyList[currNodeID] {
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

			currNodeTime := nodeAmountMap[currNodeID].min
			// check next nodes is their previous nodes complete
			// in any case mark them with time which current node achive
			for _, nextNodeID := range adjacencyListInversed[currNodeID] {
				allPrevComplete := true

				nodeAmountMap[nextNodeID].SetMin(currNodeTime - nodesValueMap[nextNodeID])
				for _, prevForNextNodeID := range adjacencyList[nextNodeID] {
					_, ok := doneNodesSet[prevForNextNodeID]
					allPrevComplete = allPrevComplete && ok
				}

				if allPrevComplete {
					nextQueue = append(nextQueue, nextNodeID)
				}
			}
		}
		for k, v := range nodeAmountMap {
			fmt.Println("\t", k, v.min, v.max)
		}
		queue = nextQueue
	}

	// collect critical paths
	var res [][]int64

	var dfs func(int64, []int64)
	dfs = func(node int64, curr []int64) {
		if amount := nodeAmountMap[node]; amount.max == 0 && amount.min == amount.max {
			slices.Reverse(curr)
			path := make([]int64, len(curr))
			copy(path, curr)
			res = append(res, path)
			return
		}

		for _, nextNode := range adjacencyListInversed[node] {
			if amount := nodeAmountMap[nextNode]; amount.min == amount.max {
				dfs(nextNode, append(curr, nextNode))
			}
		}
	}

	dfs(dummyNodeID, []int64{})

	return res, nil
}
