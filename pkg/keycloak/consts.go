package keycloak

//nolint:gosec
const (
	_getAccessTokenURL         = "/auth/realms/%s/protocol/openid-connect/token"
	_clientURL                 = "/auth/admin/realms/%s/clients"
	_clientRoleURL             = "/auth/admin/realms/%s/clients/%s/roles"
	_userURL                   = "/auth/admin/realms/%s/users"
	_clientRoleInScopeURL      = "/auth/admin/realms/%s/client-scopes/%s/scope-mappings/clients/%s"
	_sendUpdatePasswordLinkURL = "/auth/admin/realms/%s/users/%s/execute-actions-email"
	_userRoleMappingClientURL  = "/auth/admin/realms/%s/users/%s/role-mappings/clients/%s"
	_usersInRoleWithClient     = "/auth/admin/realms/%s/clients/%s/roles/%s/users"
	_scopesURL                 = "/auth/admin/realms/%s/client-scopes"

	mimeApplicationJSON = "application/json"

	_responseHeaderLocation = "Location"
)

// Email actions
const (
	UpdatePasswordAction = "UPDATE_PASSWORD"
)

// Client roles
const (
	ClientRoleAdmin  = "admin"
	ClientRoleDoctor = "doctor"
)
