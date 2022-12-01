package main

import (
	"context"
	"testing"

	temporalworker "temporal_client/internal/gen/pkg/servers/temporal_worker"
	"temporal_client/internal/gen/pkg/servers/temporal_worker/somedownstream"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowWithRealActivity(t *testing.T) {
	t.Parallel()

	// Create a test service from your service definition
	testServer := temporalworker.NewTestServer(t, context.Background(), createService, ``)
	defer testServer.Close()

	// this is mocking the downstream service that the activity ActivityWithParamAndReturn calls
	testServer.Mocks.Somedownstream.Post.MockResponse(200, map[string]string{}, &somedownstream.SomeResp{
		Msg: "hi",
	})

	// adding assertions to the activity ActivityWithParamAndReturn but still using the actual activity
	testServer.Mocks.Self.ActivityWithParamAndReturn.
		ExpectRequest(temporalworker.Param1{Msg: "hi | Executing Activity"})

	// execute workflows that you want to test
	resp, err := testServer.WorkflowWithActivities(context.Background(), temporalworker.Param1{Msg: "hi"})
	require.NoError(t, err)

	// resp is just a wrapper, you can get ID and RunID from there. resp2 is the actual payload.
	resp2, err := resp.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "hi | Activity Executed", resp2.Msg2)
}

func TestWorkflowWithMockedActivity(t *testing.T) {
	t.Parallel()

	// Create a test service from your service definition
	testServer := temporalworker.NewTestServer(t, context.Background(), createService, ``)
	defer testServer.Close()

	// adding assertions to the activity ActivityWithParamAndReturn but still using the actual activity
	testServer.Mocks.Self.ActivityWithParamAndReturn.
		ExpectRequest(temporalworker.Param1{Msg: "hi | Executing Activity"}).
		MockResponse(temporalworker.Param2{Msg2: "hiii"}, nil)

	// execute workflows that you want to test
	resp, err := testServer.WorkflowWithActivities(context.Background(), temporalworker.Param1{Msg: "hi"})
	require.NoError(t, err)

	// resp is just a wrapper, you can get ID and RunID from there. resp2 is the actual payload.
	resp2, err := resp.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "hiii | Activity Executed", resp2.Msg2)
}

func TestActivity(t *testing.T) {
	t.Parallel()
	testServer := temporalworker.NewTestServer(t, context.Background(), createService, ``)
	defer testServer.Close()

	testServer.Mocks.Somedownstream.Post.
		ExpectBody(somedownstream.SomeReq{
			Msg: "this is request",
		}).
		MockResponse(
			200,
			map[string]string{},
			&somedownstream.SomeResp{
				Msg: "hi",
			},
		)

	resp, err := testServer.ActivityWithParamAndReturn(
		context.Background(),
		temporalworker.Param1{
			Msg: "this is request",
		})
	require.NoError(t, err)
	resp2, err := resp.Get()
	require.NoError(t, err)
	assert.Equal(t, "hi", resp2.Msg2)
}
