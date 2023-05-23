package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_miscellaneous/internal/gen/pkg/servers/gateway"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/array_response_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/encoder_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/multi_contenttype_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/oneof_backend"
)

type AppConfig struct{}

func GetPingList(ctx context.Context, req *gateway.GetPingIdListRequest, client gateway.GetPingIdListClient) (*gateway.Pong, error) {
	backendReq := &encoder_backend.GetPingListRequest{
		ID: req.ID,
	}

	encoderResponse, err := client.Encoder_backendGetPingList(ctx, backendReq)
	if err != nil {
		return nil, err
	}

	return &gateway.Pong{
		Identifier: encoderResponse.Identifier,
	}, nil
}

func GetPingString(ctx context.Context, req *gateway.GetPingStringSRequest, client gateway.GetPingStringSClient) (*gateway.PongString, error) {
	backendReq := &encoder_backend.GetPingStringSRequest{
		S: req.S,
	}

	encoderResponse, err := client.Encoder_backendGetPingStringS(ctx, backendReq)
	if err != nil {
		return nil, err
	}

	return &gateway.PongString{
		S: encoderResponse.S,
	}, nil
}

func PostPingBinary(_ context.Context, req *gateway.PostPingBinaryRequest) (*gateway.GatewayBinaryResponse, error) {
	return &gateway.GatewayBinaryResponse{
		Content: req.Request.Content,
	}, nil
}

func PatchPing(_ context.Context, req *gateway.PatchPingRequest) (*gateway.GatewayPatchResponse, error) {
	return &gateway.GatewayPatchResponse{
		Content: req.Request.Content,
	}, nil
}

func PostRotateOneOf(ctx context.Context, req *gateway.PostRotateOneOfRequest, client gateway.PostRotateOneOfClient) (*gateway.OneOfResponse, error) {
	proor := &oneof_backend.PostRotateOneOfRequest{}
	for _, v := range req.Request.Values {
		switch {
		case v.One != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{One: &oneof_backend.One{One: v.One.One}})
		case v.Two != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{Two: &oneof_backend.Two{Two: v.Two.Two}})
		case v.Three != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{Three: &oneof_backend.Three{Three: v.Three.Three}})
		case v.EmptyType != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{EmptyType: &oneof_backend.EmptyType{}})
		}
	}

	resp, err := client.Oneof_backendPostRotateOneOf(ctx, proor)
	if err != nil {
		return nil, err
	}

	goor := &gateway.OneOfResponse{}
	for _, v := range resp.Values {
		switch {
		case v.One != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{One: &gateway.One{One: v.One.One}})
		case v.Two != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{Two: &gateway.Two{Two: v.Two.Two}})
		case v.Three != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{Three: &gateway.Three{Three: v.Three.Three}})
		case v.EmptyType != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{EmptyType: &gateway.EmptyType{}})
		}
	}

	return goor, nil
}

func GetPingMultiCode(_ context.Context, req *gateway.GetPingMultiCodeRequest) (*gateway.Pong, *gateway.PongString, error) {
	if req.Code == 0 {
		return &gateway.Pong{Identifier: 0}, nil, nil
	} else if req.Code == 1 {
		return nil, &gateway.PongString{S: "One"}, nil
	}

	return nil, nil, fmt.Errorf("code can only be 0 or 1")
}

//nolint:funlen
func GetPingMultiContentBackend(ctx context.Context, req *gateway.GetPingMultiContentBackendSRequest, client gateway.GetPingMultiContentBackendSClient) (*gateway.PongString, error) {
	backendReq1 := &multi_contenttype_backend.PostPingMultiColonRequest{
		Request: multi_contenttype_backend.Post_ping_multiColon_req_body_type{
			Post_ping_multiColon_req_body_type_application_json: &multi_contenttype_backend.Post_ping_multiColon_req_body_type_application_json{
				Val: &req.S,
			},
		},
	}
	backendReq2 := &multi_contenttype_backend.PostPingMultiColonRequest{
		Request: multi_contenttype_backend.Post_ping_multiColon_req_body_type{
			Post_ping_multiColon_req_body_type_application_json_Charset__Utf: &multi_contenttype_backend.Post_ping_multiColon_req_body_type_application_json_Charset__Utf{
				Val: &req.S,
			},
		},
	}
	backendReq3 := &multi_contenttype_backend.PostPingMultiUrlencodedRequest{
		Request: multi_contenttype_backend.Post_ping_multiUrlEncoded_req_body_type{
			Post_ping_multiUrlEncoded_req_body_type_application_xWwwFormUrlencoded: &multi_contenttype_backend.Post_ping_multiUrlEncoded_req_body_type_application_xWwwFormUrlencoded{
				Val: &req.S,
			},
		},
	}
	backendReq4 := &multi_contenttype_backend.PostPingMultiUrlencodedRequest{
		Request: multi_contenttype_backend.Post_ping_multiUrlEncoded_req_body_type{
			Post_ping_multiUrlEncoded_req_body_type_application_xWwwFormUrlencoded_CharsetUtf: &multi_contenttype_backend.Post_ping_multiUrlEncoded_req_body_type_application_xWwwFormUrlencoded_CharsetUtf{
				Val: &req.S,
			},
		},
	}

	encoderResponse1, err := client.Multi_contenttype_backendPostPingMultiColon(ctx, backendReq1)
	if err != nil {
		return nil, err
	}

	encoderResponse2, err := client.Multi_contenttype_backendPostPingMultiColon(ctx, backendReq2)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/x-www-form-urlencoded")
	urlEncodedCtx := common.RequestHeaderToContext(ctx, headers)

	encoderResponse3, err := client.Multi_contenttype_backendPostPingMultiUrlencoded(urlEncodedCtx, backendReq3)
	if err != nil {
		return nil, err
	}

	headers.Set("Content-Type", "application/x-www-form-urlencoded; charset = utf-8")
	urlEncodedCtx = common.RequestHeaderToContext(ctx, headers)

	encoderResponse4, err := client.Multi_contenttype_backendPostPingMultiUrlencoded(urlEncodedCtx, backendReq4)
	if err != nil {
		return nil, err
	}

	if encoderResponse1.Val == nil ||
		encoderResponse2.Val == nil ||
		encoderResponse3.Val == nil ||
		encoderResponse4.Val == nil ||
		*encoderResponse1.Val != *encoderResponse2.Val ||
		*encoderResponse1.Val != *encoderResponse3.Val ||
		*encoderResponse1.Val != *encoderResponse4.Val {
		return nil, common.CreateError(ctx, common.InternalError, "Values don't match!!!", errors.New("Values don't match!!!"))
	}

	return &gateway.PongString{
		S: *encoderResponse1.Val,
	}, nil
}

func GetPingAsync(ctx context.Context, req *gateway.GetPingAsyncdownstreamsListRequest, client gateway.GetPingAsyncdownstreamsListClient) (*gateway.Pong, error) {
	backend1Req := &encoder_backend.GetPingListRequest{
		ID: req.ID,
	}
	backend2Req := &multi_contenttype_backend.PostPingMultiColonRequest{}

	ctx1 := common.RespHeaderAndStatusToContext(ctx, make(http.Header), 0)
	ctx2 := common.RespHeaderAndStatusToContext(ctx, make(http.Header), 0)

	backend1Future := common.Async(ctx1, func(ctxInt context.Context) (interface{}, error) {
		return client.Encoder_backendGetPingList(ctxInt, backend1Req)
	})
	backend2Future := common.Async(ctx2, func(ctxInt context.Context) (interface{}, error) {
		return client.Multi_contenttype_backendPostPingMultiColon(ctxInt, backend2Req)
	})

	// above 2 calls have been made asynchronously, lets get their results
	encoderResponseInterface, err := backend1Future.Get()
	if err != nil {
		return nil, err
	}

	// for this test I am ignoring the value of the second result, as long as it's not an error
	_, err = backend2Future.Get()
	if err != nil {
		return nil, err
	}

	encoderResponse := encoderResponseInterface.(*encoder_backend.Pong)

	return &gateway.Pong{
		Identifier: encoderResponse.Identifier,
	}, nil
}

func GetEmptyResponse(_ context.Context, _ *gateway.GetEmptyResponseListRequest) (*gateway.Get_emptyResponse_200_resp_type_body, error) {
	return &gateway.Get_emptyResponse_200_resp_type_body{}, nil
}

func GetPingArrayResponseList(ctx context.Context, _ *gateway.GetPingArrayResponseListRequest, client gateway.GetPingArrayResponseListClient) ([]gateway.Res, error) {
	backendReq := &array_response_backend.GetArrayResponseListRequest{}
	backendRes, err := client.Array_response_backendGetArrayResponseList(ctx, backendReq)
	if err != nil {
		return nil, err
	}

	res := make([]gateway.Res, 0, len(backendRes))
	for _, v := range backendRes {
		res = append(res, gateway.Res{Val: v.Val})
	}

	return res, nil
}

func GetPingStringResponseList(ctx context.Context, _ *gateway.GetPingStringResponseListRequest, client gateway.GetPingStringResponseListClient) (string, error) {
	backendReq := &array_response_backend.GetStringResponseListRequest{}
	backendRes, err := client.Array_response_backendGetStringResponseList(ctx, backendReq)
	if err != nil {
		return "", err
	}

	return backendRes, nil
}

func GetPingBytesResponseList(ctx context.Context, _ *gateway.GetPingBytesResponseListRequest, client gateway.GetPingBytesResponseListClient) ([]byte, error) {
	backendReq := &array_response_backend.GetBytesResponseListRequest{}
	backendRes, err := client.Array_response_backendGetBytesResponseList(ctx, backendReq)
	if err != nil {
		return nil, err
	}

	res := make([]byte, len(backendRes))
	copy(res, backendRes)

	return res, nil
}

func GetWithHeader(_ context.Context, _ *gateway.GetWithHeaderListRequest) (*gateway.WithHeaderResponse, error) {
	return &gateway.WithHeaderResponse{}, nil
}

func createService(_ context.Context, _ AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
	return &gateway.ServiceInterface{
		PostRotateOneOf:             PostRotateOneOf,
		GetPingIdList:               GetPingList,
		GetPingStringS:              GetPingString,
		PatchPing:                   PatchPing,
		PostPingBinary:              PostPingBinary,
		GetPingMultiCode:            GetPingMultiCode,
		GetPingMultiContentBackendS: GetPingMultiContentBackend,
		GetPingAsyncdownstreamsList: GetPingAsync,
		GetEmptyResponseList:        GetEmptyResponse,
		GetPingArrayResponseList:    GetPingArrayResponseList,
		GetPingStringResponseList:   GetPingStringResponseList,
		GetPingBytesResponseList:    GetPingBytesResponseList,
		GetWithHeaderList:           GetWithHeader,
	}, nil, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx, createService)
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
