package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OAuthConfig holds provider-specific credentials loaded from config.
type OAuthConfig struct {
	GoogleClientID       string
	GoogleClientSecret   string
	GoogleRedirectURL    string
	MicrosoftClientID    string
	MicrosoftClientSecret string
	MicrosoftTenantID    string
	MicrosoftRedirectURL string
	FrontendCallbackURL  string // e.g. http://localhost:3000/auth/callback
}

// authCode is a short-lived code linking an OAuth callback to frontend token exchange.
type authCode struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// OAuthService manages OIDC provider setup, PKCE/state generation, token exchange,
// and the short-lived auth-code store used for the secure callback flow.
type OAuthService struct {
	cfg            OAuthConfig
	googleOAuth    *oauth2.Config
	microsoftOAuth *oauth2.Config
	googleProvider *oidc.Provider
	msProvider     *oidc.Provider

	mu    sync.Mutex
	codes map[string]authCode // in-memory, cleaned up by background goroutine
}

func NewOAuthService(ctx context.Context, cfg OAuthConfig) (*OAuthService, error) {
	svc := &OAuthService{
		cfg:   cfg,
		codes: make(map[string]authCode),
	}

	// Initialize Google OIDC provider
	googleProv, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("init google oidc provider: %w", err)
	}
	svc.googleProvider = googleProv
	svc.googleOAuth = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Endpoint:     googleProv.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	// Initialize Microsoft OIDC provider (tenant-specific, not /common)
	msIssuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", cfg.MicrosoftTenantID)
	msProv, err := oidc.NewProvider(ctx, msIssuer)
	if err != nil {
		return nil, fmt.Errorf("init microsoft oidc provider: %w", err)
	}
	svc.msProvider = msProv
	svc.microsoftOAuth = &oauth2.Config{
		ClientID:     cfg.MicrosoftClientID,
		ClientSecret: cfg.MicrosoftClientSecret,
		RedirectURL:  cfg.MicrosoftRedirectURL,
		Endpoint:     msProv.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	// Start background cleanup for expired auth codes
	go svc.cleanupLoop()

	return svc, nil
}

// OAuthParams holds state, nonce, and PKCE fields used during the OAuth handshake.
type OAuthParams struct {
	State        string
	Nonce        string
	CodeVerifier string
	CodeChallenge string
}

// GenerateOAuthParams creates cryptographically random state, nonce, and PKCE pair.
func GenerateOAuthParams() (OAuthParams, error) {
	state, err := randomBase64(32)
	if err != nil {
		return OAuthParams{}, err
	}
	nonce, err := randomBase64(32)
	if err != nil {
		return OAuthParams{}, err
	}
	verifier, err := randomBase64(32)
	if err != nil {
		return OAuthParams{}, err
	}

	// PKCE S256 challenge: BASE64URL(SHA256(verifier))
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	return OAuthParams{
		State:         state,
		Nonce:         nonce,
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
	}, nil
}

// AuthURL returns the provider's authorization URL with PKCE and nonce.
func (s *OAuthService) AuthURL(provider, state, nonce, challenge string) (string, error) {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", nonce),
	}
	switch provider {
	case "google":
		// hd param restricts to specific GSuite domain (additional server-side validation in callback)
		opts = append(opts, oauth2.SetAuthURLParam("hd", "hcmus.edu.vn"))
		return s.googleOAuth.AuthCodeURL(state, opts...), nil
	case "microsoft":
		return s.microsoftOAuth.AuthCodeURL(state, opts...), nil
	default:
		return "", fmt.Errorf("unknown provider: %s", provider)
	}
}

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

// IssueAuthCode stores a short-lived (60s) code mapping to a JWT pair.
// The frontend exchanges this code for tokens via POST /api/auth/oauth/exchange.
func (s *OAuthService) IssueAuthCode(accessToken, refreshToken string) string {
	code := generateHex(16)
	s.mu.Lock()
	s.codes[code] = authCode{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(60 * time.Second),
	}
	s.mu.Unlock()
	return code
}

// ConsumeAuthCode returns and deletes a stored auth code. Returns false if expired or not found.
func (s *OAuthService) ConsumeAuthCode(code string) (accessToken, refreshToken string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.codes[code]
	if !exists || time.Now().After(entry.ExpiresAt) {
		delete(s.codes, code)
		return "", "", false
	}
	delete(s.codes, code)
	return entry.AccessToken, entry.RefreshToken, true
}

// OAuthStateCookie is serialized into the httpOnly state cookie.
type OAuthStateCookie struct {
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"code_verifier"`
	Provider     string `json:"provider"`
}

// MarshalStateCookie serializes state into a JSON string for cookie storage.
func MarshalStateCookie(params OAuthParams, provider string) (string, error) {
	sc := OAuthStateCookie{
		State:        params.State,
		Nonce:        params.Nonce,
		CodeVerifier: params.CodeVerifier,
		Provider:     provider,
	}
	b, err := json.Marshal(sc)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ParseStateCookie deserializes the state cookie.
func ParseStateCookie(encoded string) (OAuthStateCookie, error) {
	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return OAuthStateCookie{}, fmt.Errorf("decode state cookie: %w", err)
	}
	var sc OAuthStateCookie
	if err := json.Unmarshal(b, &sc); err != nil {
		return OAuthStateCookie{}, fmt.Errorf("unmarshal state cookie: %w", err)
	}
	return sc, nil
}

// SetStateCookie sets the OAuth state as an httpOnly cookie on the response.
func SetStateCookie(w http.ResponseWriter, encoded string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
	})
}

// ClearStateCookie removes the OAuth state cookie.
func ClearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// FrontendCallbackURL returns the URL to redirect the browser after successful OAuth.
func (s *OAuthService) FrontendCallbackURL(code string) string {
	return s.cfg.FrontendCallbackURL + "?code=" + code
}

func (s *OAuthService) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.codes {
			if now.After(v.ExpiresAt) {
				delete(s.codes, k)
			}
		}
		s.mu.Unlock()
	}
}

func randomBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
