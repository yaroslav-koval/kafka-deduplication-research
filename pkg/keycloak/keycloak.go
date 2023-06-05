package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"net/http"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

type HTTPClient interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}

type Client struct {
	cfg        *env.KeycloakClient
	httpClient HTTPClient
}

type Request struct {
	Path        string
	Method      string
	QueryParams []*QParam
	Headers     map[string]string
	Body        interface{}
	BodyString  string
	Destination interface{}
	AccessToken string
}

type ResponseHeader struct {
	Location string
}

type QParam struct {
	Key   string
	Value string
}

type Option interface {
	apply(*Client)
}

type httpClientOption struct {
	httpClient HTTPClient
}

func (hco httpClientOption) apply(c *Client) {
	c.httpClient = hco.httpClient
}

func WithHTTPClient(httpClient HTTPClient) Option {
	return httpClientOption{httpClient: httpClient}
}

func New(cfg *env.KeycloakClient, opts ...Option) *Client {
	c := Client{
		cfg:        cfg,
		httpClient: new(fasthttp.Client),
	}

	for _, o := range opts {
		o.apply(&c)
	}

	return &c
}

func (c *Client) GetRolesByClientID(ctx context.Context, clientID, token string) ([]*ClientRoleResponse, error) {
	result := make([]*ClientRoleResponse, 0)
	r := &Request{
		Path:        fmt.Sprintf(_clientRoleURL, c.cfg.Realm, clientID),
		Method:      http.MethodGet,
		Destination: &result,
		AccessToken: token,
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetClientRoleByName(ctx context.Context, clientID, roleName, token string) (*ClientRoleResponse, error) {
	result := new(ClientRoleResponse)
	r := &Request{
		Path:        fmt.Sprintf("%s/%s", fmt.Sprintf(_clientRoleURL, c.cfg.Realm, clientID), roleName),
		Method:      http.MethodGet,
		Destination: result,
		AccessToken: token,
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetClientRoleInScopeByParams(ctx context.Context, scopeID, clientID, token string) ([]*ClientRoleResponse, error) {
	result := make([]*ClientRoleResponse, 0)
	r := &Request{
		Path:        fmt.Sprintf(_clientRoleInScopeURL, c.cfg.Realm, scopeID, clientID),
		Method:      http.MethodGet,
		Destination: &result,
		AccessToken: token,
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetUserRoleCorrespondingWithClientByParams(ctx context.Context, userID, clientID, token string) (
	[]*UserRoleCorrespondingWithClientResponse, error) {
	result := make([]*UserRoleCorrespondingWithClientResponse, 0)
	r := &Request{
		Path:        fmt.Sprintf(_userRoleMappingClientURL, c.cfg.Realm, userID, clientID),
		Method:      http.MethodGet,
		Destination: &result,
		AccessToken: token,
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetScopes(ctx context.Context, token string) ([]*GetScopesResponse, error) {
	result := make([]*GetScopesResponse, 0)

	r := &Request{
		Path:        fmt.Sprintf(_scopesURL, c.cfg.Realm),
		Method:      http.MethodGet,
		Destination: &result,
		AccessToken: token,
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetUsersInRoleWithClient(ctx context.Context, clientID, role, token string) ([]*UserResponse, error) {
	result := make([]*UserResponse, 0)
	r := &Request{
		Path:        fmt.Sprintf(_usersInRoleWithClient, c.cfg.Realm, clientID, role),
		Method:      http.MethodGet,
		Destination: &result,
		AccessToken: token,
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//nolint:dupl
func (c *Client) SearchClientByParams(ctx context.Context, params *SearchClientByParamsRequest, token string) (
	[]*SearchClientByParamsResponse, error) {
	result := make([]*SearchClientByParamsResponse, 0)
	r := &Request{
		Path:   fmt.Sprintf(_clientURL, c.cfg.Realm),
		Method: http.MethodGet,
		QueryParams: []*QParam{
			{
				Key:   "clientId",
				Value: params.ClientID,
			},
		},
		AccessToken: token,
		Destination: &result,
	}

	respHeader := &ResponseHeader{}

	err := c.sendRequest(ctx, r, respHeader)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) CreateClient(ctx context.Context, body *CreateClientRequest, token string) (*BaseCreateResourceResponse, error) {
	r := &Request{
		Path:        fmt.Sprintf(_clientURL, c.cfg.Realm),
		Method:      http.MethodPost,
		AccessToken: token,
		Body:        body,
	}

	respHeader := &ResponseHeader{}

	err := c.sendRequest(ctx, r, respHeader)
	if err != nil {
		return nil, err
	}

	return &BaseCreateResourceResponse{ResourceURL: respHeader.Location}, nil
}

func (c *Client) CreateClientRole(ctx context.Context, body *CreateClientRoleRequest, clientID, token string) error {
	r := &Request{
		Path:        fmt.Sprintf(_clientRoleURL, c.cfg.Realm, clientID),
		Method:      http.MethodPost,
		AccessToken: token,
		Body:        body,
	}

	return c.sendRequest(ctx, r, nil)
}

func (c *Client) CreateClientRoleInScope(
	ctx context.Context,
	body []*CreateClientRoleInScopeRequest,
	scopeID, clientID, token string) error {
	r := &Request{
		Path:        fmt.Sprintf(_clientRoleInScopeURL, c.cfg.Realm, scopeID, clientID),
		Method:      http.MethodPost,
		AccessToken: token,
		Body:        body,
	}

	return c.sendRequest(ctx, r, nil)
}

func (c *Client) CreateUserRoleCorrespondingWithClient(
	ctx context.Context,
	body []*CreateUserRoleCorrespondingWithClientRequest,
	userID, clientID, token string) error {
	r := &Request{
		Path:        fmt.Sprintf(_userRoleMappingClientURL, c.cfg.Realm, userID, clientID),
		Method:      http.MethodPost,
		AccessToken: token,
		Body:        body,
	}

	return c.sendRequest(ctx, r, nil)
}

//nolint:dupl
func (c *Client) SearchUserByParams(ctx context.Context, params *SearchUserRequest, token string) (
	[]*SearchUserResponse, error) {
	result := make([]*SearchUserResponse, 0)
	r := &Request{
		Path:   fmt.Sprintf(_userURL, c.cfg.Realm),
		Method: http.MethodGet,
		QueryParams: []*QParam{
			{
				Key:   "username",
				Value: params.UserName,
			},
		},
		AccessToken: token,
		Destination: &result,
	}

	respHeader := &ResponseHeader{}

	err := c.sendRequest(ctx, r, respHeader)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) DeleteUser(ctx context.Context, userID, token string) error {
	r := &Request{
		Path:        fmt.Sprintf("%s/%s", fmt.Sprintf(_userURL, c.cfg.Realm), userID),
		Method:      http.MethodDelete,
		AccessToken: token,
	}

	return c.sendRequest(ctx, r, nil)
}

func (c *Client) DeleteUserInRoleWithClient(ctx context.Context,
	body []*DeleteUserInRoleWithClientRequest, userID, clientID, token string) error {
	r := &Request{
		Path:        fmt.Sprintf(_userRoleMappingClientURL, c.cfg.Realm, userID, clientID),
		Method:      http.MethodDelete,
		Body:        body,
		AccessToken: token,
	}

	return c.sendRequest(ctx, r, nil)
}

func (c *Client) GetAccessToken(ctx context.Context) (*TokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", c.cfg.GrantType)
	params.Add("client_id", c.cfg.ClientID)
	params.Add("client_secret", c.cfg.ClientSecret)
	params.Add("username", c.cfg.Username)
	params.Add("password", c.cfg.Password)

	dst := new(TokenResponse)

	r := &Request{
		Path:        fmt.Sprintf(_getAccessTokenURL, c.cfg.Realm),
		Method:      http.MethodPost,
		BodyString:  params.Encode(),
		Destination: dst,
		Headers:     map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	}

	err := c.sendRequest(ctx, r, nil)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func (c *Client) sendRequest(ctx context.Context, r *Request, rp *ResponseHeader) error {
	var err error

	req := fasthttp.AcquireRequest()
	if r.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer "+r.AccessToken)
	}

	defer fasthttp.ReleaseRequest(req)
	req.Header.SetContentType(mimeApplicationJSON)
	req.Header.SetMethod(r.Method)

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	req.SetRequestURI(fmt.Sprintf("%s/%s", c.cfg.HostAPI, r.Path))

	for _, p := range r.QueryParams {
		req.URI().QueryArgs().Add(p.Key, p.Value)
	}

	if r.BodyString != "" {
		req.SetBodyString(r.BodyString)
	}

	if r.Body != nil {
		b, mErr := json.Marshal(r.Body)
		if mErr != nil {
			return cerror.New(ctx, cerror.KindInternal, err).LogError()
		}

		req.SetBody(b)
	}

	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(resp)

	if c.cfg.RequestTimeout != 0 {
		err = c.httpClient.DoTimeout(req, resp, c.cfg.RequestTimeout)
	} else {
		err = c.httpClient.Do(req, resp)
	}

	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	respStatus := resp.StatusCode()
	respBody := resp.Body()

	if rp != nil {
		rp.Location = string(resp.Header.Peek(_responseHeaderLocation))
	}

	if !(respStatus >= http.StatusOK && respStatus < http.StatusMultipleChoices) {
		return cerror.NewF(ctx, cerror.KindFromHTTPCode(respStatus), "keycloak unsuccessful response. Code: %d", respStatus).
			WithPayload(string(respBody)).
			LogError()
	}

	if r.Destination != nil {
		if err := json.Unmarshal(respBody, r.Destination); err != nil {
			return cerror.New(ctx, cerror.KindInternal, err).LogError()
		}
	}

	return nil
}
