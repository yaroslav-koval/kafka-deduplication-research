package keycloak

type TokenRequest struct {
	GrantType    string `json:"grant_type,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type SearchClientByParamsRequest struct {
	ClientID string `json:"clientId"`
}

type SearchClientByParamsResponse struct {
	ID string `json:"id"`
}

type CreateClientRequest struct {
	Name                    string   `json:"name,omitempty"`
	ClientID                string   `json:"clientId,omitempty"`
	Enabled                 bool     `json:"enabled"`
	StandardFlowEnabled     bool     `json:"standardFlowEnabled"`
	ImplicitFlowEnabled     bool     `json:"implicitFlowEnabled"`
	ServiceAccountsEnabled  bool     `json:"serviceAccountsEnabled"`
	PublicClient            bool     `json:"publicClient"`
	FullScopeAllowed        bool     `json:"fullScopeAllowed"`
	Protocol                string   `json:"protocol,omitempty"`
	ConsentRequired         bool     `json:"consentRequired"`
	ClientAuthenticatorType string   `json:"clientAuthenticatorType,omitempty"`
	DefaultClientScopes     []string `json:"defaultClientScopes,omitempty"`
	OptionalClientScopes    []string `json:"optionalClientScopes,omitempty"`
}

type CreateClientRoleRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
	ContainerID string `json:"containerId,omitempty"`
}

type CreateClientRoleInScopeRequest struct {
	ID string `json:"id,omitempty"`
	CreateClientRoleRequest
}

type SearchUserRequest struct {
	UserName string `json:"username"`
}

type SearchUserResponse struct {
	ID string `json:"id"`
}

type BaseCreateResourceResponse struct {
	ResourceURL string `json:"url"`
}

type CreateUserRoleCorrespondingWithClientRequest struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
	ContainerID string `json:"containerId,omitempty"`
}

type ClientRoleResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserRoleCorrespondingWithClientResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GetScopesResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserResponse struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
}

type DeleteUserInRoleWithClientRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContainerID string `json:"containerId"`
}
