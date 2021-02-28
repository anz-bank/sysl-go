package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	gateway "rest_with_conditional_downstream/internal/gen/pkg/servers/gateway"
	backend "rest_with_conditional_downstream/internal/gen/pkg/servers/gateway/backend"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
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

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
				GetFizzbuzz: GetFizzbuzz,
			}, nil, nil
		},
	)
}

func main() {
	ctx := log.PutLogger(context.Background(), log.NewDefaultLogger())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err, "something goes wrong")
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
