package grpc

import (
	"context"

	"github.com/tonikpro/grpc_test/service"
	"github.com/tonikpro/grpc_test/transport/grpc/pb"
)

type GrpcServer struct {
	pb.UnimplementedA2BillingServer
	ep service.A2billingEndpoints
}

func (s GrpcServer) GetAgentIdsByParentAgentID(ctx context.Context, request *pb.GetAgentIdsByParentAgentIDRequest) (response *pb.GetAgentIdsByParentAgentIDResponse, err error) {
	agentID := request.GetAgentId()
	r, err := s.ep.GetAgentIdsByParentAgentID(ctx, agentID)
	if err != nil {
		return
	}
	response.AgentIds = r.([]int32)
	return
}

func NewGrpcServer(ep service.A2billingEndpoints) GrpcServer {
	return GrpcServer{ep: ep}
}
