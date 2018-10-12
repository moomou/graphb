package graphb

type operationType string

// 3 types of operation.
const (
	TypeDgraphQuery  operationType = "dgraph"
	TypeQuery        operationType = "query"
	TypeMutation     operationType = "mutation"
	TypeSubscription operationType = "subscription"
)
