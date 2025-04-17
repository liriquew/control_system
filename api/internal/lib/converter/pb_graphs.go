package converter

import (
	"github.com/liriquew/control_system/internal/entities"
	"github.com/liriquew/control_system/internal/models"
	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"
)

func ConvertGraphToModel(graph *grph_pb.Graph) *models.Graph {
	return &models.Graph{
		ID:        graph.ID,
		CreatedBy: graph.CreatedBy,
		Name:      graph.Name,
		GroupID:   graph.GroupID,
	}
}

func ConvertGraphToProto(graph *models.Graph) *grph_pb.Graph {
	return &grph_pb.Graph{
		ID:        graph.ID,
		CreatedBy: graph.CreatedBy,
		Name:      graph.Name,
		GroupID:   graph.GroupID,
	}
}

func ConvertNodeToModel(node *grph_pb.Node) *models.Node {
	return &models.Node{
		ID:                node.ID,
		GraphID:           node.GraphID,
		TaskID:            node.TaskID,
		DependencyNodeIDs: node.DependencyNodeIDs,
	}
}

func ConvertNodeToProto(node *models.Node) *grph_pb.Node {
	return &grph_pb.Node{
		ID:                node.ID,
		GraphID:           node.GraphID,
		TaskID:            node.TaskID,
		DependencyNodeIDs: node.DependencyNodeIDs,
	}
}

func convertNodesToProto(nodes []*models.Node) []*grph_pb.Node {
	res := make([]*grph_pb.Node, 0, len(nodes))
	for _, node := range nodes {
		res = append(res, ConvertNodeToProto(node))
	}
	return res
}

func convertNodesToModel(nodes []*grph_pb.Node) []*models.Node {
	res := make([]*models.Node, 0, len(nodes))
	for _, node := range nodes {
		res = append(res, ConvertNodeToModel(node))
	}
	return res
}

func ConvertGraphWithNodesToProto(graph *entities.GraphWithNodes) *grph_pb.GraphWithNodes {
	return &grph_pb.GraphWithNodes{
		GraphInfo: ConvertGraphToProto(graph.GraphInfo),
		Nodes:     convertNodesToProto(graph.Nodes),
	}
}

func ConvertGraphWithNodesToModel(graph *grph_pb.GraphWithNodes) *entities.GraphWithNodes {
	return &entities.GraphWithNodes{
		GraphInfo: ConvertGraphToModel(graph.GraphInfo),
		Nodes:     convertNodesToModel(graph.Nodes),
	}
}

func ConvertGraphsWithNodesToModel(graphs []*grph_pb.GraphWithNodes) []*entities.GraphWithNodes {
	res := make([]*entities.GraphWithNodes, 0, len(graphs))
	for _, graph := range graphs {
		res = append(res, ConvertGraphWithNodesToModel(graph))
	}
	return res
}

func ConvertPredictedGraph(graph *grph_pb.PredictedGraphResponse) *entities.PredictedGraph {
	nodes := make([]*entities.NodeWithTask, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes = append(nodes, &entities.NodeWithTask{
			Node:                   ConvertNodeToModel(node.Node),
			Task:                   ConvertTaskToModel(node.Task),
			PredictedTime:          node.PredictedTime,
			AdditionalDependencies: node.AdditionalDependencies,
		})
	}

	paths := make([][]int64, 0, len(graph.Paths))
	for _, path := range graph.Paths {
		paths = append(paths, path.NodeIDs)
	}

	return &entities.PredictedGraph{
		Graph:           ConvertGraphToModel(graph.Graph),
		Nodes:           nodes,
		Paths:           paths,
		UnpredictedUIDs: graph.UnpredictedUIDs,
	}
}
