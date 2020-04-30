package repository

type A2billingRepository interface {
	GetAgentIdsByParentAgentID(id int32) ([]int32, error)
}
