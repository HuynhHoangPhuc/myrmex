package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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
