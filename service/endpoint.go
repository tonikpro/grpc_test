package service

import "github.com/go-kit/kit/endpoint"

type A2billingEndpoints struct {
	GetAgentIdsByParentAgentID endpoint.Endpoint
}
