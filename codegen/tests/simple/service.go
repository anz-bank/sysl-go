// Code generated by sysl DO NOT EDIT.
package simple

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/anz-bank/sysl-go/codegen/tests/deps"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/restlib"
	"github.com/anz-bank/sysl-go/validator"
)

// Service interface for Simple
type Service interface {
	GetApiDocsList(ctx context.Context, req *GetApiDocsListRequest) (*[]deps.ApiDoc, error)
	GetGetSomeBytesList(ctx context.Context, req *GetGetSomeBytesListRequest) (*Pdf, error)
	GetJustOkAndJustErrorList(ctx context.Context, req *GetJustOkAndJustErrorListRequest) (*http.Header, error)
	GetJustReturnErrorList(ctx context.Context, req *GetJustReturnErrorListRequest) error
	GetJustReturnOkList(ctx context.Context, req *GetJustReturnOkListRequest) (*http.Header, error)
	GetOkTypeAndJustErrorList(ctx context.Context, req *GetOkTypeAndJustErrorListRequest) (*Response, error)
	GetOopsList(ctx context.Context, req *GetOopsListRequest) (*Response, error)
	GetPetaList(ctx context.Context, req *GetPetaListRequest) (*PetA, error)
	GetPetbList(ctx context.Context, req *GetPetbListRequest) (*PetB, error)
	GetRawList(ctx context.Context, req *GetRawListRequest) (*Str, error)
	GetRawIntList(ctx context.Context, req *GetRawIntListRequest) (*Integer, error)
	GetRawStatesList(ctx context.Context, req *GetRawStatesListRequest) (*[]Status, error)
	GetRawIdStatesList(ctx context.Context, req *GetRawIdStatesListRequest) (*Str, error)
	GetRawStates2List(ctx context.Context, req *GetRawStates2ListRequest) (*Str, error)
	GetSimpleAPIDocsList(ctx context.Context, req *GetSimpleAPIDocsListRequest) (*deps.ApiDoc, error)
	GetStuffList(ctx context.Context, req *GetStuffListRequest) (*Stuff, error)
	PostStuff(ctx context.Context, req *PostStuffRequest) (*Str, error)
}

// Client for Simple API
type Client struct {
	client *http.Client
	url    string
}

// NewClient for Simple
func NewClient(client *http.Client, serviceURL string) *Client {
	return &Client{client, serviceURL}
}

// GetApiDocsList ...
func (s *Client) GetApiDocsList(ctx context.Context, req *GetApiDocsListRequest) (*[]deps.ApiDoc, error) {
	required := []string{}
	var okResponse []deps.ApiDoc
	u, err := url.Parse(fmt.Sprintf("%s/api-docs", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkDepsApiDocResponse, ok := result.Response.(*[]deps.ApiDoc)
	if ok {
		valErr := validator.Validate(OkDepsApiDocResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkDepsApiDocResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetGetSomeBytesList ...
func (s *Client) GetGetSomeBytesList(ctx context.Context, req *GetGetSomeBytesListRequest) (*Pdf, error) {
	required := []string{}
	var okResponse Pdf
	u, err := url.Parse(fmt.Sprintf("%s/get-some-bytes", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkPdfResponse, ok := result.Response.(*Pdf)
	if ok {
		valErr := validator.Validate(OkPdfResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkPdfResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetJustOkAndJustErrorList ...
func (s *Client) GetJustOkAndJustErrorList(ctx context.Context, req *GetJustOkAndJustErrorListRequest) (*http.Header, error) {
	required := []string{}
	u, err := url.Parse(fmt.Sprintf("%s/just-ok-and-just-error", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, nil, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	return &result.HTTPResponse.Header, nil
}

// GetJustReturnErrorList ...
func (s *Client) GetJustReturnErrorList(ctx context.Context, req *GetJustReturnErrorListRequest) error {
	required := []string{}
	u, err := url.Parse(fmt.Sprintf("%s/just-return-error", s.url))
	if err != nil {
		return common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, nil, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	return nil
}

// GetJustReturnOkList ...
func (s *Client) GetJustReturnOkList(ctx context.Context, req *GetJustReturnOkListRequest) (*http.Header, error) {
	required := []string{}
	u, err := url.Parse(fmt.Sprintf("%s/just-return-ok", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, nil, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	return &result.HTTPResponse.Header, nil
}

// GetOkTypeAndJustErrorList ...
func (s *Client) GetOkTypeAndJustErrorList(ctx context.Context, req *GetOkTypeAndJustErrorListRequest) (*Response, error) {
	required := []string{}
	var okResponse Response
	u, err := url.Parse(fmt.Sprintf("%s/ok-type-and-just-error", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkResponseResponse, ok := result.Response.(*Response)
	if ok {
		valErr := validator.Validate(OkResponseResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkResponseResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetOopsList ...
func (s *Client) GetOopsList(ctx context.Context, req *GetOopsListRequest) (*Response, error) {
	required := []string{}
	var okResponse Response
	var errorResponse Status
	u, err := url.Parse(fmt.Sprintf("%s/oops", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, &errorResponse)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		response, ok := err.(*restlib.HTTPResult)
		if !ok {
			return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
		}
		return nil, common.CreateDownstreamError(ctx, common.DownstreamResponseError, response.HTTPResponse, response.Body, &errorResponse)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkResponseResponse, ok := result.Response.(*Response)
	if ok {
		valErr := validator.Validate(OkResponseResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkResponseResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetPetaList ...
func (s *Client) GetPetaList(ctx context.Context, req *GetPetaListRequest) (*PetA, error) {
	required := []string{}
	var okResponse PetA
	u, err := url.Parse(fmt.Sprintf("%s/petA", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	q := u.Query()
	q.Add("id", req.ID)

	u.RawQuery = q.Encode()
	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkPetAResponse, ok := result.Response.(*PetA)
	if ok {
		valErr := validator.Validate(OkPetAResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkPetAResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetPetbList ...
func (s *Client) GetPetbList(ctx context.Context, req *GetPetbListRequest) (*PetB, error) {
	required := []string{}
	var okResponse PetB
	u, err := url.Parse(fmt.Sprintf("%s/petB", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	q := u.Query()
	q.Add("id", req.ID)

	u.RawQuery = q.Encode()
	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkPetBResponse, ok := result.Response.(*PetB)
	if ok {
		valErr := validator.Validate(OkPetBResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkPetBResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetRawList ...
func (s *Client) GetRawList(ctx context.Context, req *GetRawListRequest) (*Str, error) {
	required := []string{}
	var okResponse Str
	u, err := url.Parse(fmt.Sprintf("%s/raw", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	q := u.Query()
	q.Add("bt", fmt.Sprintf("%v", req.Bt))

	u.RawQuery = q.Encode()
	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkStrResponse, ok := result.Response.(*Str)
	if ok {
		valErr := validator.Validate(OkStrResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkStrResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetRawIntList ...
func (s *Client) GetRawIntList(ctx context.Context, req *GetRawIntListRequest) (*Integer, error) {
	required := []string{}
	var okResponse Integer
	u, err := url.Parse(fmt.Sprintf("%s/raw-int", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkIntegerResponse, ok := result.Response.(*Integer)
	if ok {
		valErr := validator.Validate(OkIntegerResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkIntegerResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetRawStatesList ...
func (s *Client) GetRawStatesList(ctx context.Context, req *GetRawStatesListRequest) (*[]Status, error) {
	required := []string{}
	var okResponse []Status
	u, err := url.Parse(fmt.Sprintf("%s/raw/states", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkStatusResponse, ok := result.Response.(*[]Status)
	if ok {
		valErr := validator.Validate(OkStatusResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkStatusResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetRawIdStatesList ...
func (s *Client) GetRawIdStatesList(ctx context.Context, req *GetRawIdStatesListRequest) (*Str, error) {
	required := []string{}
	var okResponse Str
	u, err := url.Parse(fmt.Sprintf("%s/raw/%v/states", s.url, req.ID))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkStrResponse, ok := result.Response.(*Str)
	if ok {
		valErr := validator.Validate(OkStrResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkStrResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetRawStates2List ...
func (s *Client) GetRawStates2List(ctx context.Context, req *GetRawStates2ListRequest) (*Str, error) {
	required := []string{}
	var okResponse Str
	u, err := url.Parse(fmt.Sprintf("%s/raw/%v/states2", s.url, req.ID))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkStrResponse, ok := result.Response.(*Str)
	if ok {
		valErr := validator.Validate(OkStrResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkStrResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetSimpleAPIDocsList ...
func (s *Client) GetSimpleAPIDocsList(ctx context.Context, req *GetSimpleAPIDocsListRequest) (*deps.ApiDoc, error) {
	required := []string{}
	var okResponse deps.ApiDoc
	u, err := url.Parse(fmt.Sprintf("%s/simple-api-docs", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkDepsApiDocResponse, ok := result.Response.(*deps.ApiDoc)
	if ok {
		valErr := validator.Validate(OkDepsApiDocResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkDepsApiDocResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// GetStuffList ...
func (s *Client) GetStuffList(ctx context.Context, req *GetStuffListRequest) (*Stuff, error) {
	required := []string{}
	var okResponse Stuff
	u, err := url.Parse(fmt.Sprintf("%s/stuff", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	q := u.Query()
	q.Add("it", fmt.Sprintf("%v", req.It))

	if req.Dt != nil {
		q.Add("dt", fmt.Sprintf("%v", *req.Dt))
	}

	if req.St != nil {
		q.Add("st", *req.St)
	}

	if req.Bt != nil {
		q.Add("bt", fmt.Sprintf("%v", *req.Bt))
	}

	u.RawQuery = q.Encode()
	result, err := restlib.DoHTTPRequest(ctx, s.client, "GET", u.String(), nil, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- GET "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkStuffResponse, ok := result.Response.(*Stuff)
	if ok {
		valErr := validator.Validate(OkStuffResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkStuffResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}

// PostStuff ...
func (s *Client) PostStuff(ctx context.Context, req *PostStuffRequest) (*Str, error) {
	required := []string{}
	var okResponse Str
	u, err := url.Parse(fmt.Sprintf("%s/stuff", s.url))
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "failed to parse url", err)
	}

	result, err := restlib.DoHTTPRequest(ctx, s.client, "POST", u.String(), req.Request, required, &okResponse, nil)
	restlib.OnRestResultHTTPResult(ctx, result, err)
	if err != nil {
		return nil, common.CreateError(ctx, common.DownstreamUnavailableError, "call failed: Simple <- POST "+u.String(), err)
	}

	if result.HTTPResponse.StatusCode == http.StatusUnauthorized {
		return nil, common.CreateDownstreamError(ctx, common.DownstreamUnauthorizedError, result.HTTPResponse, result.Body, nil)
	}
	OkStrResponse, ok := result.Response.(*Str)
	if ok {
		valErr := validator.Validate(OkStrResponse)
		if valErr != nil {
			return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, valErr)
		}

		return OkStrResponse, nil
	}

	return nil, common.CreateDownstreamError(ctx, common.DownstreamUnexpectedResponseError, result.HTTPResponse, result.Body, nil)
}
