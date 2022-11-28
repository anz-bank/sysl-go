package temporal_tester

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
)

type MockedWorker interface {
	GetWorkflows() []MockWorkflow
	GetName() string
}

func MockClientsBuilder(mocks ...MockedWorker) func(context.Context, string, *client.Options) (client.Client, error) {
	envs := map[string]*testsuite.TestWorkflowEnvironment{}
	for _, m := range mocks {
		envs[m.GetName()] = NewEnvWithWorkflows(m.GetWorkflows()...)
	}
	return func(ctx context.Context, s string, o *client.Options) (client.Client, error) {
		env, hasEnv := envs[s]
		if !hasEnv {
			panic(fmt.Sprintf("mock environment for %q does not exist", s))
		}
		return NewTemporalMockClient(env), nil
	}
}

func EmptyClientsBuilder(context.Context, string, *client.Options) (client.Client, error) {
	suite := &testsuite.WorkflowTestSuite{}
	return NewTemporalMockClient(suite.NewTestWorkflowEnvironment()), nil
}
