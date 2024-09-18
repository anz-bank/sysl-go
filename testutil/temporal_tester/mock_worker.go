package temporal_tester

import (
	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var _ worker.Worker = &MockWorker{}

type MockWorker struct {
	actEnv     *testsuite.TestActivityEnvironment
	mockClient *MockClient
}

func MockWorkerBuilder(client client.Client, taskQueue string, options worker.Options) worker.Worker {
	mockClient, is := client.(*MockClient)
	if !is {
		panic("MockWorker needs MockClient")
	}
	suite := &testsuite.WorkflowTestSuite{}
	mockClient.Env = mockClient.Env.SetWorkerOptions(options)

	return &MockWorker{
		actEnv:     suite.NewTestActivityEnvironment().SetWorkerOptions(options),
		mockClient: mockClient,
	}
}

func (mw *MockWorker) Start() error                             { return nil }
func (mw *MockWorker) Run(interruptCh <-chan interface{}) error { return nil }
func (mw *MockWorker) Stop()                                    {}

func (mw *MockWorker) RegisterWorkflow(w interface{}) {
	mw.mockClient.Env.RegisterWorkflow(w)
}

func (mw *MockWorker) RegisterWorkflowWithOptions(w interface{}, options workflow.RegisterOptions) {
	mw.mockClient.Env.RegisterWorkflowWithOptions(w, options)
}

func (mw *MockWorker) RegisterActivity(a interface{}) {
	mw.mockClient.Env.RegisterActivity(a)
}

func (mw *MockWorker) RegisterActivityWithOptions(a interface{}, options activity.RegisterOptions) {
	mw.mockClient.Env.RegisterActivityWithOptions(a, options)
}

func (mw *MockWorker) RegisterNexusService(s *nexus.Service) {
	mw.mockClient.Env.RegisterNexusService(s)
}

func (mw *MockWorker) GetTestActivityEnv() *testsuite.TestActivityEnvironment {
	return mw.actEnv
}

// TestFuture is a wrapper of the test result of executing test activity.
type TestFuture[Resp any] struct {
	converter.EncodedValue
}

func (t *TestFuture[Resp]) Get() (Resp, error) {
	var v Resp
	err := t.EncodedValue.Get(&v)
	return v, err
}

func ExecuteTestActivity[Resp any](env *testsuite.TestActivityEnvironment, name string, params ...any) (*TestFuture[Resp], error) {
	r, err := env.ExecuteActivity(name, params...)
	return &TestFuture[Resp]{r}, err
}
