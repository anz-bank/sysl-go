package temporal_tester

import (
	"context"
	"fmt"

	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

var (
	_ client.Client      = &MockClient{}
	_ client.WorkflowRun = &MockWorkflowRun{}
)

func NewMockWorkflowBase[Resp any](t syslgo.TestingT) *MockBase[Resp] {
	return &MockBase[Resp]{
		t: t,
	}
}

type MockBase[Resp any] struct {
	t         syslgo.TestingT
	requests  []any
	opt       client.StartWorkflowOptions
	mocked    bool
	expecting bool
	mockResp  Resp
	mockErr   error
}

func (m *MockBase[Resp]) ExpectRequest(req ...any) {
	m.requests = req
	m.expecting = true
}
func (m *MockBase[Resp]) ExpectOptions(opt client.StartWorkflowOptions) {
	m.opt = opt
	m.expecting = true
}
func (m *MockBase[Resp]) IsMocked() bool { return m.mocked }
func (m *MockBase[Resp]) MockErrorOnlyResponse(err error) {
	m.mockErr = err
	m.mocked = true
}
func (m *MockBase[Resp]) MockResponse(resp Resp, err error) {
	m.mockResp = resp
	m.mockErr = err
	m.mocked = true
}

func (m *MockBase[Resp]) CheckExpectations(actual ...any) {
	if m.requests == nil || !m.expecting {
		return
	}

	if len(actual) != len(m.requests) {
		panic("actual and expected requests are of different length")
	}

	for i, expected := range m.requests {
		if !assert.Equal(m.t, expected, actual[i], fmt.Sprintf("expected: %v, actual: %v", expected, actual[i])) {
			m.t.FailNow()
		}
	}
}

func (m *MockBase[Resp]) BuildMockWorkflow() func(workflow.Context, ...any) (Resp, error) {
	return mockWithResponse[workflow.Context](m)
}
func (m *MockBase[Resp]) BuildMockWorkflowWithoutReturn() func(workflow.Context, ...any) error {
	return mockWithoutResponse[workflow.Context](m)
}

func (m *MockBase[Resp]) BuildMockActivity() func(context.Context, ...any) (Resp, error) {
	return mockWithResponse[context.Context](m)
}

func (m *MockBase[Resp]) BuildMockActivityWithoutReturn() func(context.Context, ...any) error {
	return mockWithoutResponse[context.Context](m)
}

func mockWithResponse[Context, Resp any](m *MockBase[Resp]) func(ctx Context, a ...any) (Resp, error) {
	return func(ctx Context, a ...any) (Resp, error) {
		m.CheckExpectations(a...)
		return m.mockResp, m.mockErr
	}
}

func mockWithoutResponse[Context, Resp any](m *MockBase[Resp]) func(ctx Context, a ...any) error {
	return func(ctx Context, a ...any) error {
		m.CheckExpectations(a...)
		return m.mockErr
	}
}

type MockWorkflowRun struct {
	Env *testsuite.TestWorkflowEnvironment
	ID  string
}

func (m *MockWorkflowRun) GetID() string {
	return m.ID
}

func (m *MockWorkflowRun) GetRunID() string {
	return ""
}

func (m *MockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return m.Env.GetWorkflowResult(valuePtr)
}

func (m *MockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	return m.Env.GetWorkflowResult(valuePtr)
}

type MockClient struct {
	Env *testsuite.TestWorkflowEnvironment
}

type MockWorkflow struct {
	Workflow any
	Option   workflow.RegisterOptions
}

func NewEnvWithWorkflows(wfs ...MockWorkflow) *testsuite.TestWorkflowEnvironment {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	for _, wf := range wfs {
		env.RegisterWorkflowWithOptions(wf.Workflow, wf.Option)
	}
	return env
}

func NewTemporalMockClient(env *testsuite.TestWorkflowEnvironment) client.Client {
	return &MockClient{env}
}

func (m *MockClient) GetEnv() *testsuite.TestWorkflowEnvironment {
	return m.Env
}

func (m *MockClient) ExecuteWorkflow(
	ctx context.Context,
	options client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) (client.WorkflowRun, error) {
	if options.ID == "" {
		// need to know ID name for getting results
		options.ID = uuid.New()
	}
	m.Env = m.Env.SetStartWorkflowOptions(options)
	m.Env.ExecuteWorkflow(workflow, args...)
	return &MockWorkflowRun{m.Env, options.ID}, nil
}

func (m *MockClient) GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun {
	return nil
}

func (m *MockClient) SignalWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
	signalName string,
	arg interface{},
) error {
	return m.Env.SignalWorkflowByID(workflowID, signalName, arg)
}

func (m *MockClient) SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{}, options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error) {
	return nil, nil
}

func (m *MockClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	m.Env.CancelWorkflow()
	return nil
}

func (m *MockClient) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error {
	return nil
}

func (m *MockClient) GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enumspb.HistoryEventFilterType) client.HistoryEventIterator {
	return nil
}

func (m *MockClient) CompleteActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error {
	return nil
}

func (m *MockClient) CompleteActivityByID(ctx context.Context, namespace, workflowID, runID, activityID string, result interface{}, err error) error {
	return nil
}

func (m *MockClient) RecordActivityHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error {
	return nil
}

func (m *MockClient) RecordActivityHeartbeatByID(ctx context.Context, namespace, workflowID, runID, activityID string, details ...interface{}) error {
	return nil
}

func (m *MockClient) ListClosedWorkflow(ctx context.Context, request *workflowservice.ListClosedWorkflowExecutionsRequest) (*workflowservice.ListClosedWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (m *MockClient) ListOpenWorkflow(ctx context.Context, request *workflowservice.ListOpenWorkflowExecutionsRequest) (*workflowservice.ListOpenWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (m *MockClient) ListWorkflow(ctx context.Context, request *workflowservice.ListWorkflowExecutionsRequest) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (m *MockClient) ListArchivedWorkflow(ctx context.Context, request *workflowservice.ListArchivedWorkflowExecutionsRequest) (*workflowservice.ListArchivedWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (m *MockClient) ScanWorkflow(ctx context.Context, request *workflowservice.ScanWorkflowExecutionsRequest) (*workflowservice.ScanWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (m *MockClient) CountWorkflow(ctx context.Context, request *workflowservice.CountWorkflowExecutionsRequest) (*workflowservice.CountWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (m *MockClient) GetSearchAttributes(ctx context.Context) (*workflowservice.GetSearchAttributesResponse, error) {
	return nil, nil
}

func (m *MockClient) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	return m.Env.QueryWorkflowByID(workflowID, queryType, args...)
}

func (m *MockClient) QueryWorkflowWithOptions(ctx context.Context, request *client.QueryWorkflowWithOptionsRequest) (*client.QueryWorkflowWithOptionsResponse, error) {
	return nil, nil
}

func (m *MockClient) DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	return nil, nil
}

func (m *MockClient) DescribeTaskQueue(ctx context.Context, taskqueue string, taskqueueType enumspb.TaskQueueType) (*workflowservice.DescribeTaskQueueResponse, error) {
	return nil, nil
}

func (m *MockClient) ResetWorkflowExecution(ctx context.Context, request *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error) {
	return nil, nil
}

func (m *MockClient) CheckHealth(ctx context.Context, request *client.CheckHealthRequest) (*client.CheckHealthResponse, error) {
	return nil, nil
}

func (m *MockClient) WorkflowService() workflowservice.WorkflowServiceClient {
	return nil
}

func (m *MockClient) OperatorService() operatorservice.OperatorServiceClient {
	return nil
}

func (m *MockClient) ScheduleClient() client.ScheduleClient {
	return nil
}

func (m *MockClient) Close() {}
