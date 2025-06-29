package graphtools

type GraphWithNodes interface {
	Len() int
	GetNodes() []Node
}

type Node interface {
	GetDependencies() []int64
	GetID() int64
	GetWeight() float64
	GetAssignedTo() int64
	GetAdditionalDependencies() []int64
	AddAdditionalDependency(...int64)
}
