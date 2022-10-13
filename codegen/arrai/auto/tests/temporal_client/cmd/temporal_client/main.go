package main

import (
	"context"
	"log"

	frontdoor "temporal_client/internal/gen/pkg/servers/temporal_client"
	"temporal_client/internal/gen/pkg/servers/temporal_client/temporalworker"
	pb "temporal_client/protos"

	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct {
	// Define app-level config fields here.
}

func main() {

	frontdoor.Serve(context.Background(),
		func(ctx context.Context, config AppConfig) (*frontdoor.GrpcServiceInterface, *core.Hooks, error) {
			// Perform one-time setup based on config here.
			return &frontdoor.GrpcServiceInterface{
				Executor: func(ctx context.Context, req *pb.Req, c frontdoor.ExecutorClient) (*pb.Resp, error) {
					_, err := c.TemporalworkerWorkflowWithoutParam(ctx)
					if err != nil {
						log.Println(err)
					}
					_, err = c.TemporalworkerWorkflowWithOneParam(ctx, temporalworker.Param1{
						Msg: "executing workflow with one param",
					})
					if err != nil {
						log.Println(err)
					}

					_, err = c.TemporalworkerWorkflowWithMultipleParams(
						ctx,
						temporalworker.Param1{
							Msg: "executing workflow with multiple params, this is param 1",
						},
						temporalworker.Param2{
							Msg2: "executing workflow with multiple params, this is param 2",
						},
						temporalworker.Param3{
							Msg3: "executing workflow with multiple params, this is param 3",
						},
					)
					if err != nil {
						log.Println(err)
					}
					r, err := c.TemporalworkerWorkflowWithParamAndReturn(ctx, temporalworker.Param1{
						Msg: "executing workflow with param and return",
					})
					if err != nil {
						return nil, err
					}
					s, err := r.Get(ctx)
					if err != nil {
						log.Println(err)
					} else {
						log.Println(s)
					}

					r2, err := c.TemporalworkerProtoReqAndResp(ctx, pb.Req{
						EncoderId: "1",
						Content:   "executing workflow with proto request and response",
					})
					if err != nil {
						return nil, err
					}
					s2, err := r2.Get(ctx)
					if err != nil {
						log.Println(err)
					} else {
						log.Println(s2.String())
					}

					return &pb.Resp{
						Content: "all workflows are executed",
					}, nil
				},
			}, nil, nil
		},
	)
}
