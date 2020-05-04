package repository

//go:generate mockgen -source=a2billing.go -destination=../test/mocks/repository/a2billing.go -package=repository

type A2billingRepository interface {
	GetAgentIdsByParentAgentID(id int32) ([]int32, error)
}
