package core

import "context"

type ServiceDefinition[AppConfig, ServiceIntf any] func(context.Context, AppConfig) (ServiceIntf, *Hooks, error)
