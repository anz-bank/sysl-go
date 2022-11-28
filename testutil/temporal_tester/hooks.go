package temporal_tester

import (
	"context"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/stretchr/testify/require"
)

// TODO: use this on other test services.
func PatchedService[
	AppConfig, ServiceInterface any,
](
	t syslgo.TestingT,
	createService core.ServiceDefinition[AppConfig, ServiceInterface],
	testHooks *core.Hooks,
	withActualDownstreams bool,
) func(context.Context, AppConfig) (ServiceInterface, *core.Hooks, error) {
	return func(ctx context.Context, ac AppConfig) (ServiceInterface, *core.Hooks, error) {
		svc, hooks, err := createService(ctx, ac)
		require.NoError(t, err)

		hooks.ShouldSetGrpcGlobalLogger = testHooks.ShouldSetGrpcGlobalLogger
		hooks.HTTPClientBuilder = testHooks.HTTPClientBuilder
		hooks.StoppableServerBuilder = testHooks.StoppableServerBuilder
		hooks.StoppableGrpcServerBuilder = testHooks.StoppableGrpcServerBuilder
		hooks.ValidateConfig = testHooks.ValidateConfig

		if !withActualDownstreams {
			hooks.ExperimentalTemporalClientBuilder = EmptyClientsBuilder
			hooks.ExperimentalTemporalWorkerBuilder = MockWorkerBuilder
		}

		return svc, hooks, nil
	}
}
