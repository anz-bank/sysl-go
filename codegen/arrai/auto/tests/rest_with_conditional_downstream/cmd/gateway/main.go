package main

import (
	"context"
	"fmt"
	"strings"

	gateway "rest_with_conditional_downstream/internal/gen/pkg/servers/gateway"
	backend "rest_with_conditional_downstream/internal/gen/pkg/servers/gateway/backend"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct{}

func GetFizzbuzz(ctx context.Context, req *gateway.GetFizzbuzzRequest, client gateway.GetFizzbuzzClient) (*gateway.GatewayResponse, error) {
	var b strings.Builder
	var i int64
	for i = 1; i <= req.N; i++ {
		var response *backend.Response
		var err error
		if i%3 == 0 && i%5 != 0 {
			response, err = client.BackendPostFizzWithArg(ctx, &backend.PostFizzWithArgRequest{N: i})
		}
		if i%3 != 0 && i%5 == 0 {
			response, err = client.BackendPostBuzzWithArg(ctx, &backend.PostBuzzWithArgRequest{N: i})
		}
		if i%3 == 0 && i%5 == 0 {
			response, err = client.BackendPostFizzbuzzWithArg(ctx, &backend.PostFizzbuzzWithArgRequest{N: i})
		}
		if err != nil {
			return nil, err
		}
		if response == nil {
			continue
		}
		_, err = b.WriteString(fmt.Sprintf("%s\n", response.Content))
		if err != nil {
			return nil, err
		}
	}
	return &gateway.GatewayResponse{Content: b.String()}, nil
}

func application(ctx context.Context) {
	gateway.Serve(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			// FIXME auto codegen and common.MapError don't align.
			mapError := func(ctx context.Context, err error) *common.HTTPError {
				httpErr := common.MapError(ctx, err)
				return &httpErr
			}

			return &gateway.ServiceInterface{
					GetFizzbuzz: GetFizzbuzz,
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
	ctx := log.WithLogger(logger).Onto(context.Background())

	application(ctx)
}
