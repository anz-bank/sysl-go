package main

import (
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	frontdoor "temporal_client/internal/gen/pkg/servers/temporal_client/temporalworker"
	tq "temporal_client/internal/gen/pkg/servers/temporal_client/temporalworker"

	pb "temporal_client/protos"
)

func main() {
	// Create the client object just once per process
	c, err := client.NewClient(client.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, frontdoor.TemporalWorkerTaskQueue, worker.Options{})
	w.RegisterWorkflow(WorkflowWithoutParam)
	w.RegisterWorkflow(WorkflowWithOneParam)
	w.RegisterWorkflow(WorkflowWithMultipleParams)
	w.RegisterWorkflow(WorkflowWithParamAndReturn)
	w.RegisterWorkflow(ProtoReqAndResp)
	w.RegisterActivity(Activity)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}

func Activity(name string) (string, error) {
	greeting := fmt.Sprintf("Hello %s!", name)
	return greeting, nil
}

var actOpt = workflow.ActivityOptions{
	StartToCloseTimeout: time.Second * 5,
	TaskQueue:           frontdoor.TemporalWorkerTaskQueue,
}

type Param struct {
	Msg string `json:"msg" url:"msg"`
}

type Param2 struct {
	Msg2 string `json:"msg2" url:"msg2"`
}

type Param3 struct {
	Msg3 string `json:"msg3" url:"msg3"`
}

type Resp struct {
	Msg2 string `json:"msg2" url:"msg2"`
}

func WorkflowWithoutParam(ctx workflow.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, actOpt)
	var result string
	err := workflow.ExecuteActivity(ctx, Activity, "workflow without param").Get(ctx, &result)
	return result, err
}

func WorkflowWithParamAndReturn(ctx workflow.Context, param tq.Param1) (tq.Param2, error) {
	ctx = workflow.WithActivityOptions(ctx, actOpt)
	var result string
	err := workflow.ExecuteActivity(ctx, Activity, fmt.Sprintf("workflow with param and return: %s", param.Msg)).Get(ctx, &result)
	return tq.Param2{Msg2: result}, err
}

func WorkflowWithOneParam(ctx workflow.Context, req Param) error {
	ctx = workflow.WithActivityOptions(ctx, actOpt)
	var result string
	return workflow.ExecuteActivity(ctx, Activity, fmt.Sprintf("workflow with one param: %s", req.Msg)).Get(ctx, &result)
}

func WorkflowWithMultipleParams(ctx workflow.Context, req1 Param, req2 Param2, req3 Param3) error {
	ctx = workflow.WithActivityOptions(ctx, actOpt)
	var result string
	return workflow.ExecuteActivity(ctx, Activity, fmt.Sprintf("workflow with multiple param: %q %q %q", req1.Msg, req2.Msg2, req3.Msg3)).Get(ctx, &result)
}

func ProtoReqAndResp(ctx workflow.Context, req pb.Req) (pb.Resp, error) {
	ctx = workflow.WithActivityOptions(ctx, actOpt)
	var result string
	err := workflow.ExecuteActivity(ctx, Activity, fmt.Sprintf("workflow with proto req and resp: %s", req.Content)).Get(ctx, &result)
	return pb.Resp{
		Content: result,
	}, err
}
