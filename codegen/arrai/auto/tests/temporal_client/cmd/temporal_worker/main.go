package main

import (
	"context"
	"fmt"

	"github.com/anz-bank/sysl-go/core"
	"go.temporal.io/sdk/workflow"

	temporalworker "temporal_client/internal/gen/pkg/servers/temporal_worker"
	"temporal_client/internal/gen/pkg/servers/temporal_worker/somedownstream"
)

type AppConfig struct {
	// Define app-level config fields here.
}

// FIXME: expose client.Options and worker.Options

func main() {
	temporalworker.Serve(context.Background(),
		func(ctx context.Context, config AppConfig) (*temporalworker.TemporalServiceInterface, *core.Hooks, error) {
			// Perform one-time setup based on config here.
			return &temporalworker.TemporalServiceInterface{
				WorkflowWithActivities:     WorkflowWithActivities,
				ActivityWithParamAndReturn: ActivityWithParamAndReturn,
			}, &core.Hooks{}, nil
		},
	)
}

func WorkflowWithActivities(
	ctx workflow.Context,
	activities temporalworker.WorkflowWithActivitiesActivities,
	req temporalworker.Param1,
) (temporalworker.Param2, error) {
	f := activities.ActivityWithParamAndReturn(ctx, temporalworker.Param1{
		Msg: fmt.Sprintf("%s | Executing Activity", req.Msg),
	})
	s, err := f.Get(ctx)
	if err != nil {
		return temporalworker.Param2{}, nil
	}
	return temporalworker.Param2{
		Msg2: fmt.Sprintf("%s | Activity Executed", s),
	}, nil
}

func ActivityWithParamAndReturn(
	ctx context.Context,
	client temporalworker.ActivityWithParamAndReturnClient,
	req temporalworker.Param1,
) (temporalworker.Param2, error) {
	x, err := client.SomedownstreamPost(ctx, &somedownstream.PostRequest{
		Request: somedownstream.SomeReq{
			Msg: "HIIII",
		},
	})
	if err != nil {
		return temporalworker.Param2{}, err
	}
	return temporalworker.Param2{
		Msg2: x.Msg,
	}, nil
}
