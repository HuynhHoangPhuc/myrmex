package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/query"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/domain/repository"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
)

// OAuthHandler handles the OAuth login initiation and callback for Google and Microsoft.
// Flow: /login → [redirect to provider] → /callback → issue short-lived code → frontend exchange.
type OAuthHandler struct {
	oauthSvc *auth.OAuthService
	jwtSvc   *auth.JWTService
	userRepo repository.UserRepository
}

func NewOAuthHandler(
	oauthSvc *auth.OAuthService,
	jwtSvc *auth.JWTService,
	userRepo repository.UserRepository,
) *OAuthHandler {
	return &OAuthHandler{
		oauthSvc: oauthSvc,
		jwtSvc:   jwtSvc,
		userRepo: userRepo,
	}
}

// GoogleLogin initiates Google OAuth: generates PKCE+state, sets cookie, redirects.
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	h.initiateLogin(c, "google")
}

// GoogleCallback handles the Google OAuth callback.
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	h.handleCallback(c, "google")
}

// MicrosoftLogin initiates Microsoft OAuth: generates PKCE+state, sets cookie, redirects.
func (h *OAuthHandler) MicrosoftLogin(c *gin.Context) {
	h.initiateLogin(c, "microsoft")
}

// MicrosoftCallback handles the Microsoft OAuth callback.
func (h *OAuthHandler) MicrosoftCallback(c *gin.Context) {
	h.handleCallback(c, "microsoft")
}

// ExchangeAuthCode exchanges a short-lived code for a JWT pair.
// POST /api/auth/oauth/exchange  { "code": "..." }
func (h *OAuthHandler) ExchangeAuthCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	accessToken, refreshToken, ok := h.oauthSvc.ConsumeAuthCode(req.Code)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired auth code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

// initiateLogin generates PKCE+state, persists them in a cookie, and redirects to provider.
func (h *OAuthHandler) initiateLogin(c *gin.Context, provider string) {
	params, err := auth.GenerateOAuthParams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate oauth params"})
		return
	}

	cookieVal, err := auth.MarshalStateCookie(params, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal state"})
		return
	}

	secure := c.Request.TLS != nil
	auth.SetStateCookie(c.Writer, cookieVal, secure)

	authURL, err := h.oauthSvc.AuthURL(provider, params.State, params.Nonce, params.CodeChallenge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// handleCallback validates the OAuth callback, upserts the user, issues JWT, and
// redirects to the frontend with a short-lived exchange code (tokens never in URL).
func (h *OAuthHandler) handleCallback(c *gin.Context, provider string) {
	// Read and clear state cookie
	cookieRaw, err := c.Cookie("oauth_state")
	if err != nil {
		h.redirectError(c, "missing oauth state cookie")
		return
	}
	auth.ClearStateCookie(c.Writer)

	sc, err := auth.ParseStateCookie(cookieRaw)
	if err != nil || sc.Provider != provider {
		h.redirectError(c, "invalid oauth state cookie")
		return
	}

	// Validate state parameter (CSRF protection)
	if c.Query("state") != sc.State {
		h.redirectError(c, "state mismatch")
		return
	}

	// Handle provider error responses
	if errParam := c.Query("error"); errParam != "" {
		h.redirectError(c, fmt.Sprintf("provider error: %s", errParam))
		return
	}

	code := c.Query("code")
	if code == "" {
		h.redirectError(c, "missing auth code")
		return
	}

	ctx := c.Request.Context()

	// Exchange code → verify ID token → extract user info
	info, err := h.oauthSvc.ExchangeAndVerify(ctx, provider, code, sc.CodeVerifier, sc.Nonce)
	if err != nil {
		h.redirectError(c, "token verification failed: "+err.Error())
		return
	}

	// Determine role from email domain
	role := roleFromEmail(info.Email)
	if role == "" {
		h.redirectError(c, "unauthorized email domain")
		return
	}

	// Find or create user (requires pre-existing teacher/student record)
	user, err := h.userRepo.GetByOAuth(ctx, provider, info.Subject)
	if err != nil {
		h.redirectError(c, "user lookup failed")
		return
	}

	if user == nil {
		// No existing OAuth link → try to find by email (pre-created account)
		user, err = h.userRepo.GetByEmail(ctx, info.Email)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			h.redirectError(c, "user lookup failed")
			return
		}
		if user == nil {
			// No account at all → reject (admin must pre-create the teacher/student)
			h.redirectError(c, "no account found — contact admin to set up your account")
			return
		}
	}

	// Upsert: link OAuth identity to existing or new account
	user, err = h.userRepo.UpsertOAuthUser(ctx,
		info.Email, info.Name, string(user.Role),
		provider, info.Subject, info.Picture,
	)
	if err != nil {
		h.redirectError(c, "failed to link oauth account")
		return
	}

	if !user.IsActive {
		h.redirectError(c, "account is inactive")
		return
	}

	// Resolve dept/teacher scope for JWT
	deptID, teacherID := query.ResolveTokenClaims(ctx, user, h.userRepo)

	accessToken, err := h.jwtSvc.GenerateAccessToken(user.ID.String(), string(user.Role), deptID, teacherID)
	if err != nil {
		h.redirectError(c, "token generation failed")
		return
	}
	refreshToken, err := h.jwtSvc.GenerateRefreshToken(user.ID.String())
	if err != nil {
		h.redirectError(c, "token generation failed")
		return
	}

	// Issue short-lived code; redirect browser to frontend — tokens never appear in URL
	authCode := h.oauthSvc.IssueAuthCode(accessToken, refreshToken)
	c.Redirect(http.StatusFound, h.oauthSvc.FrontendCallbackURL(authCode))
}

// roleFromEmail determines the expected role based on email domain.
// Returns "" for unrecognized domains (rejected by handler).
func roleFromEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	switch parts[1] {
	case "hcmus.edu.vn":
		return "teacher"
	case "student.hcmus.edu.vn":
		return "student"
	default:
		return ""
	}
}

// redirectError redirects the browser to the frontend login page with an error message.
func (h *OAuthHandler) redirectError(c *gin.Context, msg string) {
	// Redirect to frontend login with error param so user sees a friendly message
	c.Redirect(http.StatusFound, h.oauthSvc.FrontendCallbackURL("error:"+msg))
}
