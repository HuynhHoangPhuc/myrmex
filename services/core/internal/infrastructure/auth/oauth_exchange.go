package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// UserInfo holds the verified identity returned from a provider's ID token.
type UserInfo struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

// ExchangeAndVerify exchanges the auth code for tokens and validates the ID token.
// Returns verified user claims from the provider.
func (s *OAuthService) ExchangeAndVerify(ctx context.Context, provider, code, verifier, nonce string) (UserInfo, error) {
	switch provider {
	case "google":
		return s.exchangeGoogle(ctx, code, verifier, nonce)
	case "microsoft":
		return s.exchangeMicrosoft(ctx, code, verifier, nonce)
	default:
		return UserInfo{}, fmt.Errorf("unknown provider: %s", provider)
	}
}

func (s *OAuthService) exchangeGoogle(ctx context.Context, code, verifier, nonce string) (UserInfo, error) {
	token, err := s.googleOAuth.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return UserInfo{}, fmt.Errorf("google token exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return UserInfo{}, fmt.Errorf("google: no id_token in response")
	}

	verifierOIDC := s.googleProvider.Verifier(&oidc.Config{ClientID: s.cfg.GoogleClientID})
	idToken, err := verifierOIDC.Verify(ctx, rawIDToken)
	if err != nil {
		return UserInfo{}, fmt.Errorf("google id_token verify: %w", err)
	}

	var claims struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		HD      string `json:"hd"`    // hosted domain (Google Workspace)
		Nonce   string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return UserInfo{}, fmt.Errorf("google claims: %w", err)
	}

	// Server-side hosted domain validation — must be @hcmus.edu.vn
	if claims.HD != "hcmus.edu.vn" {
		return UserInfo{}, fmt.Errorf("google: unauthorized domain %q (expected hcmus.edu.vn)", claims.HD)
	}
	if claims.Nonce != nonce {
		return UserInfo{}, fmt.Errorf("google: nonce mismatch")
	}

	return UserInfo{
		Subject: idToken.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	}, nil
}

func (s *OAuthService) exchangeMicrosoft(ctx context.Context, code, verifier, nonce string) (UserInfo, error) {
	token, err := s.microsoftOAuth.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return UserInfo{}, fmt.Errorf("microsoft token exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return UserInfo{}, fmt.Errorf("microsoft: no id_token in response")
	}

	verifierOIDC := s.msProvider.Verifier(&oidc.Config{ClientID: s.cfg.MicrosoftClientID})
	idToken, err := verifierOIDC.Verify(ctx, rawIDToken)
	if err != nil {
		return UserInfo{}, fmt.Errorf("microsoft id_token verify: %w", err)
	}

	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
		TenantID          string `json:"tid"`
		Nonce             string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return UserInfo{}, fmt.Errorf("microsoft claims: %w", err)
	}

	if claims.TenantID != s.cfg.MicrosoftTenantID {
		return UserInfo{}, fmt.Errorf("microsoft: unauthorized tenant %q", claims.TenantID)
	}
	if claims.Nonce != nonce {
		return UserInfo{}, fmt.Errorf("microsoft: nonce mismatch")
	}

	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}

	return UserInfo{
		Subject: idToken.Subject,
		Email:   email,
		Name:    claims.Name,
	}, nil
}
