package middleware

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleClaims struct {
	Subject string
	Email   string
	Name    string
}

func VerifyGoogleToken(ctx context.Context, tokenString, clientID string) (*GoogleClaims, error) {
	payload, err := idtoken.Validate(ctx, tokenString, clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)

	return &GoogleClaims{
		Subject: payload.Subject,
		Email:   email,
		Name:    name,
	}, nil
}
