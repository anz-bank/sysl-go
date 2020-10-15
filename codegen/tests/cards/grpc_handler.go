// Code generated by sysl DO NOT EDIT.
package cards

import (
	"context"

	pb "github.com/anz-bank/sysl-go/codegen/tests/cardspb"
	"github.com/anz-bank/sysl-go/core"
	"google.golang.org/grpc"
)

// GrpcServiceHandler for Cards API
type GrpcServiceHandler struct {
	genCallback      core.GrpcGenCallback
	serviceInterface *GrpcServiceInterface
	unimpl           *pb.UnimplementedCardsServer
}

// NewGrpcServiceHandler for Cards
func NewGrpcServiceHandler(genCallback core.GrpcGenCallback, serviceInterface *GrpcServiceInterface) *GrpcServiceHandler {
	return &GrpcServiceHandler{genCallback, serviceInterface, &(pb.UnimplementedCardsServer{})}
}

// RegisterServer registers the Cards gRPC service
func (s *GrpcServiceHandler) RegisterServer(ctx context.Context, server *grpc.Server) {
	pb.RegisterCardsServer(server, s)
}

// GetCards ...
func (s *GrpcServiceHandler) GetCards(ctx context.Context, req *pb.GetCardsRequest) (*pb.GetCardsResponse, error) {
	if s.serviceInterface.GetCards == nil {
		return s.unimpl.GetCards(ctx, req)
	}

	ctx, cancel := s.genCallback.DownstreamTimeoutContext(ctx)
	defer cancel()
	client := GetCardsClient{}

	return s.serviceInterface.GetCards(ctx, req, client)
}