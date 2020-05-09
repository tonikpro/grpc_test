package repository

import "context"

//go:generate mockgen -source=a2billing.go -destination=../test/mocks/repository/a2billing.go -package=repository

type A2billingRepository interface {
	GetAgentIdsByParentAgentID(context.Context, int32) ([]int32, error)
}
