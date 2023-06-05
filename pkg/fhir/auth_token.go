package fhir

import (
	"context"
	"kafka-polygon/pkg/cerror"

	"github.com/golang-jwt/jwt"
)

type JWTClient struct {
	parser *jwt.Parser
}

func NewJWTClient() *JWTClient {
	return &JWTClient{
		parser: new(jwt.Parser),
	}
}

type TokenClaims struct {
	RequestOrganizationID string `json:"azp"`
}

func (c *TokenClaims) Valid() error {
	return nil
}

func (j *JWTClient) ParseUnverified(ctx context.Context, t string) (*TokenClaims, error) {
	tc := new(TokenClaims)

	_, _, err := j.parser.ParseUnverified(t, tc)
	if err != nil {
		return nil, cerror.NewF(ctx, cerror.KindPermission, "failed to parse token: %s", err.Error()).LogError()
	}

	return tc, nil
}
