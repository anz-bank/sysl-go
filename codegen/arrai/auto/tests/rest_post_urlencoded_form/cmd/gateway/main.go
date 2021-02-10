package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"

	gateway "rest_post_urlencoded_form/internal/gen/pkg/servers/gateway"
	bananastand "rest_post_urlencoded_form/internal/gen/pkg/servers/gateway/bananastand"
)

type AppConfig struct{}

func PostBanana(ctx context.Context, req *gateway.PostBananaRequest, client gateway.PostBananaClient) (*gateway.GatewayResponse, error) {
	// unpack request
	tokens := strings.Split(req.Request.Content, ":")
	if len(tokens) != 2 {
		return nil, errors.New("bad request")
	}
	// POST form data to external banana stand service
	header := make(http.Header)
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx = common.RequestHeaderToContext(ctx, header)
	bananaStandReq := &bananastand.PostBananaRequest{
		Request: bananastand.BananaRequest{
			Client_id:     tokens[0],
			Client_secret: tokens[1],
		},
	}
	bananaStandResponse, err := client.BananastandPostBanana(ctx, bananaStandReq)
	if err != nil {
		return nil, err
	}
	// convert banana stand response back to gateway response
	if bananaStandResponse.Banana == nil {
		return nil, errors.New("banana not found")
	}
	return &gateway.GatewayResponse{
		Content: *bananaStandResponse.Banana,
	}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
				PostBanana: PostBanana,
			}, nil, nil
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
