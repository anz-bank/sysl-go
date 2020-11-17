package main

import (
	"context"
	"fmt"
	"os"

	gateway "rest_error_downstream/internal/gen/pkg/servers/gateway"
	backend "rest_error_downstream/internal/gen/pkg/servers/gateway/backend"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct{}

// FIXME why does autogen tack a "List" suffix on the name of this?
func GetApiDoopList(ctx context.Context, req *gateway.GetApiDoopListRequest, client gateway.GetApiDoopListClient) (*gateway.GatewayResponse, error) {
	err := client.BackendPostDoop(ctx, &backend.PostDoopRequest{})
	if err != nil {
		downstreamErr, ok := err.(*common.DownstreamError)
		if ok {
			ter, ok := downstreamErr.Cause.(*backend.ErrorResponse)
			if ok {
				// Note: sysl-go autogen secretly name-mangles the field with a "_" suffix
				// to disambiguate it from the Error() method.
				msg := fmt.Sprintf("backend sent us an ErrorResponse: %s", ter.Error_)
				return &gateway.GatewayResponse{Content: msg}, nil
			}
		}
		return nil, err
	}
	return &gateway.GatewayResponse{Content: "???"}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {

			// FIXME auto codegen and common.MapError don't align.
			mapError := func(ctx context.Context, err error) *common.HTTPError {
				httpErr := common.MapError(ctx, err)
				return &httpErr
			}

			return &gateway.ServiceInterface{
					GetApiDoopList: GetApiDoopList,
				}, &core.Hooks{
					MapError: mapError,
				},
				nil
		},
	)
}

func main() {
	// initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).WithConfigs(log.SetVerboseMode(true)).Onto(context.Background())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err)
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
