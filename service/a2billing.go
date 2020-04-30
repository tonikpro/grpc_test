package service

import "context"

type A2billing interface {
	GetAgentIdsByParentAgentID(ctx context.Context, id int32) ([]int32, error)
}
