package keycloak_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/keycloak"
	"kafka-polygon/pkg/log"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type clientTestSuite struct {
	suite.Suite
	keycloakClient    *keycloak.Client
	keycloakClientCfg *env.KeycloakClient
	httpClient        *mockClient
}

type mockClient struct {
	DoFunc func(req *fasthttp.Request, resp *fasthttp.Response) error
}

func (m *mockClient) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return m.DoFunc(req, resp)
}

func (m *mockClient) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return m.DoFunc(req, resp)
}

func TestClientTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(clientTestSuite))
}

func (s *clientTestSuite) SetupSuite() {
	s.httpClient = new(mockClient)
	s.keycloakClientCfg = &env.KeycloakClient{
		HostAPI:        "http://hostapi",
		Realm:          "ksa-ehealth",
		RequestTimeout: 30,
		GrantType:      "password",
		ClientID:       "client_id",
		ClientSecret:   "client_secret",
		Username:       "username",
		Password:       "password",
	}
	s.keycloakClient = keycloak.New(s.keycloakClientCfg, keycloak.WithHTTPClient(s.httpClient))
}

func (s *clientTestSuite) TearDownTest() {
}

func (s *clientTestSuite) TearDownSuite() {
}

func (s *clientTestSuite) TestGetRolesByClientID() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/clients/%s/roles", s.keycloakClientCfg.HostAPI, clientID), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.ClientRoleResponse{{ID: "id", Name: "name"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.GetRolesByClientID(ctx, clientID.String(), "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
	s.Equal("name", r[0].Name)
}

func (s *clientTestSuite) TestGetClientRoleByName() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	roleName := "admin"
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/clients/%s/roles/%s",
			s.keycloakClientCfg.HostAPI, clientID, roleName), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := &keycloak.ClientRoleResponse{ID: "id", Name: "name"}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.GetClientRoleByName(ctx, clientID.String(), "admin", "token")
	s.NoError(err)

	s.Equal("id", r.ID)
	s.Equal("name", r.Name)
}

//nolint:dupl
func (s *clientTestSuite) TestGetClientRoleInScopeByParams() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	scopeID := uuid.NewV4()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/client-scopes/%s/scope-mappings/clients/%s",
			s.keycloakClientCfg.HostAPI, scopeID, clientID), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.ClientRoleResponse{{ID: "id", Name: "name"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.GetClientRoleInScopeByParams(ctx, scopeID.String(), clientID.String(), "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
	s.Equal("name", r[0].Name)
}

//nolint:dupl
func (s *clientTestSuite) TestGetUserRoleCorrespondingWithClientByParams() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	userID := uuid.NewV4()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/users/%s/role-mappings/clients/%s",
			s.keycloakClientCfg.HostAPI, userID, clientID), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.UserRoleCorrespondingWithClientResponse{{ID: "id", Name: "name"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.GetUserRoleCorrespondingWithClientByParams(ctx, userID.String(), clientID.String(), "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
	s.Equal("name", r[0].Name)
}

func (s *clientTestSuite) TestGetScopes() {
	ctx := context.Background()

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/client-scopes",
			s.keycloakClientCfg.HostAPI), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.GetScopesResponse{{ID: "id", Name: "name"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.GetScopes(ctx, "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
	s.Equal("name", r[0].Name)
}

//nolint:dupl
func (s *clientTestSuite) TestGetUsersInRoleWithClient() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	role := "prt0001"
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/clients/%s/roles/%s/users",
			s.keycloakClientCfg.HostAPI, clientID, role), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.UserResponse{{ID: "id", UserName: "username"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.GetUsersInRoleWithClient(ctx, clientID.String(), "prt0001", "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
	s.Equal("username", r[0].UserName)
}

//nolint:dupl
func (s *clientTestSuite) TestSearchUserByParams() {
	ctx := context.Background()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/users?username=user", s.keycloakClientCfg.HostAPI), req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.SearchUserResponse{{ID: "id"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.SearchUserByParams(ctx, &keycloak.SearchUserRequest{UserName: "user"}, "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
}

//nolint:dupl
func (s *clientTestSuite) TestSearchClientByParams() {
	ctx := context.Background()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(
			fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/clients?clientId=123", s.keycloakClientCfg.HostAPI),
			req.URI().String())
		s.Equal(fasthttp.MethodGet, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		res := []*keycloak.SearchClientByParamsResponse{{ID: "id"}}
		b, err := json.Marshal(res)
		s.NoError(err)

		_, err = resp.BodyWriter().Write(b)
		s.NoError(err)

		return nil
	}

	r, err := s.keycloakClient.SearchClientByParams(ctx, &keycloak.SearchClientByParamsRequest{ClientID: "123"}, "token")
	s.NoError(err)

	s.Equal("id", r[0].ID)
}

func (s *clientTestSuite) TestCreateClient() {
	ctx := context.Background()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/clients", s.keycloakClientCfg.HostAPI), req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		dst := &keycloak.CreateClientRequest{}
		err := json.Unmarshal(req.Body(), dst)
		s.NoError(err)
		s.Equal("name", dst.Name)
		s.Equal("client id", dst.ClientID)
		s.Equal(true, dst.Enabled)
		s.Equal(true, dst.StandardFlowEnabled)
		s.Equal(true, dst.ImplicitFlowEnabled)
		s.Equal(true, dst.ServiceAccountsEnabled)
		s.Equal(true, dst.PublicClient)
		s.Equal(true, dst.FullScopeAllowed)
		s.Equal("protocol", dst.Protocol)
		s.Equal(true, dst.ConsentRequired)
		s.Equal("client authenticator type", dst.ClientAuthenticatorType)
		s.Equal([]string{"default"}, dst.DefaultClientScopes)
		s.Equal([]string{"optional"}, dst.OptionalClientScopes)

		resp.Header.Add("Location", "client id")

		return nil
	}

	r, err := s.keycloakClient.CreateClient(ctx, &keycloak.CreateClientRequest{
		Name:                    "name",
		ClientID:                "client id",
		Enabled:                 true,
		StandardFlowEnabled:     true,
		ImplicitFlowEnabled:     true,
		ServiceAccountsEnabled:  true,
		PublicClient:            true,
		FullScopeAllowed:        true,
		Protocol:                "protocol",
		ConsentRequired:         true,
		ClientAuthenticatorType: "client authenticator type",
		DefaultClientScopes:     []string{"default"},
		OptionalClientScopes:    []string{"optional"},
	}, "token")
	s.NoError(err)
	s.Equal("client id", r.ResourceURL)
}

func (s *clientTestSuite) TestCreateClientRole() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(
			fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/clients/%s/roles",
				s.keycloakClientCfg.HostAPI, clientID.String()), req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		dst := &keycloak.CreateClientRoleRequest{}
		err := json.Unmarshal(req.Body(), dst)
		s.NoError(err)
		s.Equal("name", dst.Name)
		s.Equal("description", dst.Description)
		s.Equal(true, dst.Composite)
		s.Equal(true, dst.ClientRole)
		s.Equal("container id", dst.ContainerID)

		return nil
	}

	err := s.keycloakClient.CreateClientRole(ctx, &keycloak.CreateClientRoleRequest{
		Name:        "name",
		Description: "description",
		Composite:   true,
		ClientRole:  true,
		ContainerID: "container id",
	}, clientID.String(), "token")
	s.NoError(err)
}

func (s *clientTestSuite) TestCreateClientRoleInScope() {
	ctx := context.Background()
	clientID := uuid.NewV4()
	scopeID := uuid.NewV4()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/client-scopes/%s/scope-mappings/clients/%s",
			s.keycloakClientCfg.HostAPI, scopeID, clientID), req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		dst := []*keycloak.CreateClientRoleInScopeRequest{}
		err := json.Unmarshal(req.Body(), &dst)
		s.NoError(err)
		s.Equal("id", dst[0].ID)
		s.Equal("name", dst[0].Name)
		s.Equal("description", dst[0].Description)
		s.Equal(true, dst[0].Composite)
		s.Equal(true, dst[0].ClientRole)
		s.Equal("container id", dst[0].ContainerID)

		return nil
	}

	err := s.keycloakClient.CreateClientRoleInScope(ctx, []*keycloak.CreateClientRoleInScopeRequest{
		{
			ID: "id",
			CreateClientRoleRequest: keycloak.CreateClientRoleRequest{
				Name:        "name",
				Description: "description",
				Composite:   true,
				ClientRole:  true,
				ContainerID: "container id",
			},
		},
	}, scopeID.String(), clientID.String(), "token")
	s.NoError(err)
}

func (s *clientTestSuite) TestCreateUserRoleCorrespondingWithClient() {
	ctx := context.Background()
	userID := uuid.NewV4()
	clientID := uuid.NewV4()
	expDst := []*keycloak.CreateUserRoleCorrespondingWithClientRequest{
		{
			ID:          "id",
			Name:        "name",
			Description: "description",
			Composite:   true,
			ClientRole:  true,
			ContainerID: "container id",
		},
	}

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/users/%s/role-mappings/clients/%s",
			s.keycloakClientCfg.HostAPI, userID.String(), clientID.String()), req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		expB, _ := json.Marshal(expDst)
		s.Equal(string(expB), string(req.Body()))

		return nil
	}

	err := s.keycloakClient.CreateUserRoleCorrespondingWithClient(ctx, expDst, userID.String(), clientID.String(), "token")
	s.NoError(err)
}

func (s *clientTestSuite) TestDeleteUser() {
	ctx := context.Background()
	userID := uuid.NewV4()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/users/%s",
			s.keycloakClientCfg.HostAPI, userID.String()), req.URI().String())
		s.Equal(fasthttp.MethodDelete, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		return nil
	}

	err := s.keycloakClient.DeleteUser(ctx, userID.String(), "token")
	s.NoError(err)
}

func (s *clientTestSuite) TestDeleteUserInRoleWithClient() {
	ctx := context.Background()
	userID := uuid.NewV4()
	clientID := uuid.NewV4()
	expBody := []*keycloak.DeleteUserInRoleWithClientRequest{{
		ID:          "id",
		Name:        "name",
		ContainerID: "container id",
	}}

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		s.Equal(fmt.Sprintf("%s/auth/admin/realms/ksa-ehealth/users/%s/role-mappings/clients/%s",
			s.keycloakClientCfg.HostAPI, userID.String(), clientID), req.URI().String())
		s.Equal(fasthttp.MethodDelete, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		basicAuthPrefix := []byte("Bearer ")
		auth := req.Header.Peek("Authorization")
		s.Equal(true, bytes.HasPrefix(auth, basicAuthPrefix))

		token := string(auth[len(basicAuthPrefix):])
		s.Equal("token", token)

		s.Equal("application/json", string(req.Header.Peek("content-type")))

		expB, _ := json.Marshal(expBody)
		s.Equal(string(expB), string(req.Body()))

		return nil
	}

	err := s.keycloakClient.DeleteUserInRoleWithClient(ctx, expBody, userID.String(), clientID.String(), "token")
	s.NoError(err)
}

func (s *clientTestSuite) TestGetAccessToken() {
	ctx := context.Background()
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		url := fmt.Sprintf("%s/auth/realms/ksa-ehealth/protocol/openid-connect/token", s.keycloakClientCfg.HostAPI)
		s.Equal(url, req.URI().String())
		s.Equal(fasthttp.MethodPost, string(req.Header.Method()))
		resp.SetStatusCode(fasthttp.StatusOK)

		s.Equal("application/x-www-form-urlencoded", string(req.Header.Peek("content-type")))

		dst := string(req.Body())
		s.Equal("client_id=client_id&client_secret=client_secret&grant_type=password&password=password&username=username", dst)

		respBody := &keycloak.TokenResponse{
			AccessToken:      "token",
			ExpiresIn:        60,
			RefreshExpiresIn: 60,
			RefreshToken:     "refresh token",
			TokenType:        "type",
			NotBeforePolicy:  1,
			SessionState:     "session state",
			Scope:            "scope",
		}
		b, _ := json.Marshal(respBody)
		_, _ = resp.BodyWriter().Write(b)

		return nil
	}

	t, err := s.keycloakClient.GetAccessToken(ctx)
	s.NoError(err)
	s.Equal("token", t.AccessToken)
	s.Equal(60, t.ExpiresIn)
	s.Equal(60, t.RefreshExpiresIn)
	s.Equal("refresh token", t.RefreshToken)
	s.Equal("type", t.TokenType)
	s.Equal(1, t.NotBeforePolicy)
	s.Equal("session state", t.SessionState)
	s.Equal("scope", t.Scope)
}
