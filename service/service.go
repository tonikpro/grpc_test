package service

import (
	"context"

	"github.com/tonikpro/grpc_test/repository"
)

type a2billing struct {
	repo repository.A2billingRepository
}

func (s a2billing) GetAgentIdsByParentAgentID(ctx context.Context, id int32) (result []int32, err error) {
	return s.repo.GetAgentIdsByParentAgentID(ctx, id)
}

func NewService(repo repository.A2billingRepository) A2billing {
	return &a2billing{repo}
}
