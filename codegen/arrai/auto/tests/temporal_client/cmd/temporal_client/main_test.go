package main

import (
	"context"
	"errors"
	frontdoor "temporal_client/internal/gen/pkg/servers/temporal_client"
	"temporal_client/internal/gen/pkg/servers/temporal_client/temporalworker"
	pb "temporal_client/protos"
	"testing"
)

func TestExecutorMock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testServer := frontdoor.NewTestServer(t, ctx, createService, ``)
	defer testServer.Close()

	testServer.Mocks.Temporalworker.WorkflowWithActivities.MockResponse(
		temporalworker.Param2{
			Msg2: "this is a mock",
		},
		nil,
	)

	testServer.Executor().
		WithContext(ctx).
		WithRequest(&pb.Req{
			EncoderId: "test",
			Content:   "test",
		}).
		ExpectResponse(&pb.Resp{
			Content: "all workflows are executed this is a mock",
		}).
		Send()
}

func TestExecutorMockError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testServer := frontdoor.NewTestServer(t, ctx, createService, ``)
	defer testServer.Close()

	testServer.Mocks.Temporalworker.WorkflowWithActivities.MockResponse(
		temporalworker.Param2{},
		errors.New("this is an error"),
	)

	testServer.Executor().
		WithContext(ctx).
		WithRequest(&pb.Req{
			EncoderId: "test",
			Content:   "test",
		}).
		ExpectError(errors.New("rpc error: code = Unknown desc = workflow execution error (type: WorkflowWithActivities, workflowID: Some Custom ID, runID: default-test-run-id): this is an error")).
		Send()
}

func TestExecutorMockWithAssertions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testServer := frontdoor.NewTestServer(t, ctx, createService, ``)
	defer testServer.Close()

	testServer.Mocks.Temporalworker.WorkflowWithActivities.
		ExpectRequest(temporalworker.Param1{
			Msg: "executing activity from client: test",
		}).
		MockResponse(temporalworker.Param2{
			Msg2: "this is another mock",
		}, nil)

	testServer.Executor().
		WithContext(ctx).
		WithRequest(&pb.Req{
			EncoderId: "test",
			Content:   "test",
		}).
		ExpectResponse(&pb.Resp{
			Content: "all workflows are executed this is another mock",
		}).
		Send()
}
