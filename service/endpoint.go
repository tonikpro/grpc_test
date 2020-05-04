package service

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

type A2billingEndpoints struct {
	GetAgentIdsByParentAgentID endpoint.Endpoint
}

func NewA2billingEndpoints(svc A2billing) A2billingEndpoints {
	ep := A2billingEndpoints{}
	ep.GetAgentIdsByParentAgentID = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		if id, ok := request.(int32); ok {
			return svc.GetAgentIdsByParentAgentID(ctx, id)
		}
		return nil, errors.New("ERR_BADREQUEST")
	}
	return ep
}
