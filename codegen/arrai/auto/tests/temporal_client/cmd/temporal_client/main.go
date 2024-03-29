package main

import (
	"context"
	"fmt"

	frontdoor "temporal_client/internal/gen/pkg/servers/temporal_client"
	"temporal_client/internal/gen/pkg/servers/temporal_client/somedownstream"
	"temporal_client/internal/gen/pkg/servers/temporal_client/temporalworker"
	pb "temporal_client/protos"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"go.temporal.io/sdk/client"
)

type AppConfig struct {
	// Define app-level config fields here.
}

func createService(ctx context.Context, config AppConfig) (*frontdoor.GrpcServiceInterface, *core.Hooks, error) {
	// Perform one-time setup based on config here.
	return &frontdoor.GrpcServiceInterface{
		Rest: func(ctx context.Context, req *pb.Req, client frontdoor.RestClient) (*pb.Resp, error) {
			r, err := client.SomedownstreamPost(ctx, &somedownstream.PostRequest{
				Request: somedownstream.SomeReq{
					Msg: "HI",
				},
			})
			if err != nil {
				return nil, err
			}
			return &pb.Resp{
				Content: r.Msg,
			}, nil
		},
		Executor: func(ctx context.Context, req *pb.Req, c frontdoor.ExecutorClient) (*pb.Resp, error) {
			r, err := c.TemporalworkerWorkflowWithActivities(ctx, temporalworker.Param1{
				Msg: fmt.Sprintf("executing activity from client: %s", req.Content),
			}, client.StartWorkflowOptions{
				ID: "Some Custom ID",
			})
			if err != nil {
				return nil, err
			}
			r2, err := r.Get(ctx)
			if err != nil {
				return nil, err
			}

			// How to get client for TemporalWorker
			_, err = common.TemporalClientFrom(ctx, temporalworker.TaskQueue)
			if err != nil {
				return nil, err
			}
			return &pb.Resp{
				Content: "all workflows are executed" + " " + r2.Msg2,
			}, nil
		},
	}, nil, nil
}

func main() {
	frontdoor.Serve(context.Background(), createService)
}
