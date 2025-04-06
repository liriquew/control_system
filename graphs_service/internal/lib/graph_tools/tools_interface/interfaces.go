package graph_tools_interface

type GraphWithNodes interface {
	Len() int
	GetNodes() []Node
}

type Node interface {
	GetDependencies() []int64
	GetID() int64
	GetAssignedTo() int64
	GetAdditionalDependencies() []int64
	AddAdditionalDependency(...int64)
}
