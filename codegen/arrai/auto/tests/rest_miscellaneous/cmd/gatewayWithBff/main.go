package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff"
	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff/gateway"
	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff/types"
)

type AppConfig struct{}

func PostPingBinary(_ context.Context, req *gatewayWithBff.PostPingBinaryRequest) (*gatewayWithBff.GatewayBinaryResponse, error) {
	return &gatewayWithBff.GatewayBinaryResponse{
		Content: req.Request.Content,
	}, nil
}

type ErrorResponseWriter struct {
	code int
	err  error
}

func (e ErrorResponseWriter) Error() string {
	return e.err.Error()
}

func (e ErrorResponseWriter) WriteError(ctx context.Context, w http.ResponseWriter) bool {
	b, err := json.Marshal(e.err)
	if err != nil {
		log.Error(ctx, err, "error marshalling error response")
		return false
	}

	w.WriteHeader(e.code)

	// Ignore write error, if any, as it is probably a client issue.
	_, _ = w.Write(b)

	return true
}

func PostMultiResponses(
	ctx context.Context, req *gatewayWithBff.PostMultiResponsesRequest, client gatewayWithBff.PostMultiResponsesClient,
) (*types.SomethingExternal, error) {
	respPong, respPongString, err := client.GatewayGetPingMultiCode(ctx, &gateway.GetPingMultiCodeRequest{Code: req.Request.Code})
	if err != nil {
		if downstreamError, ok := err.(*common.DownstreamError); ok {
			if downstreamError.Response != nil {
				switch {
				case downstreamError.Response.StatusCode == 400:
					if gatewayBinaryRequest, ok := downstreamError.Cause.(*gateway.GatewayBinaryRequest); ok {
						return nil, ErrorResponseWriter{
							code: 400,
							err:  gatewayWithBff.GatewayBinaryRequest{Content: gatewayBinaryRequest.Content}}
					}
				case downstreamError.Response.StatusCode == 500:
					if gatewayBinaryResponse, ok := downstreamError.Cause.(*gateway.GatewayBinaryResponse); ok {
						return nil, ErrorResponseWriter{
							code: 500,
							err:  gatewayWithBff.GatewayBinaryResponse{Content: gatewayBinaryResponse.Content}}
					}
				}
			}
		}
		return nil, err
	}
	if respPong != nil {
		return &types.SomethingExternal{
			Data: fmt.Sprintf("Pong response: %+v", respPong),
		}, nil
	}
	if respPongString != nil {
		return &types.SomethingExternal{
			Data: fmt.Sprintf("PongString response: %+v", respPongString.S),
		}, nil
	}
	return nil, fmt.Errorf("no responses")
}

func PostMultiStatuses(ctx context.Context, req *gatewayWithBff.PostMultiStatusesRequest, client gatewayWithBff.PostMultiStatusesClient) (*types.SomethingExternal, error) {
	respPong, respPongString, err := client.GatewayGetPingMultiCodeTypesList(ctx, &gateway.GetPingMultiCodeTypesListRequest{
		Code: req.Request.Code,
	})
	if err != nil {
		if downstreamError, ok := err.(*common.DownstreamError); ok {
			if downstreamError.Response != nil {
				switch downstreamError.Response.StatusCode {
				case 400, 500:
					if gatewayBinaryRequest, ok := downstreamError.Cause.(*gateway.GatewayBinaryRequest); ok { return nil, ErrorResponseWriter{ code: 400,
							err:  gatewayWithBff.GatewayBinaryRequest{Content: gatewayBinaryRequest.Content}}
					}
				}
			}
		}
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	if respPong != nil {
		return &types.SomethingExternal{
			Data: fmt.Sprintf("Pong response: %+v", respPong),
		}, nil
	}
	if respPongString != nil {
		return &types.SomethingExternal{
			Data: fmt.Sprintf("PongString response: %+v", respPongString.S),
		}, nil
	}
	return nil, fmt.Errorf("no responses")
}

func createService(_ context.Context, _ AppConfig) (*gatewayWithBff.ServiceInterface, *core.Hooks, error) {
	return &gatewayWithBff.ServiceInterface{
		PostMultiResponses: PostMultiResponses,
		PostMultiStatuses:  PostMultiStatuses,
		PostPingBinary:     PostPingBinary,
	}, nil, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gatewayWithBff.NewServer(ctx, createService)
}

func main() {
	ctx := log.PutLogger(context.Background(), log.NewDefaultLogger())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err, "something went wrong")
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
