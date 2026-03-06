package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OAuthConfig holds provider-specific credentials loaded from config.
type OAuthConfig struct {
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	MicrosoftClientID     string
	MicrosoftClientSecret string
	MicrosoftTenantID     string
	MicrosoftRedirectURL  string
	FrontendCallbackURL   string // e.g. http://localhost:3000/auth/callback
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
	State         string
	Nonce         string
	CodeVerifier  string
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
