package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/auth"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/infrastructure/llm"
)

// authServices groups the JWT, password, and optional OAuth service instances.
type authServices struct {
	JWT      *auth.JWTService
	Password *auth.PasswordService
	OAuth    *auth.OAuthService // nil when OAuth is not configured
}

// initAuthServices creates JWT, password, and optionally OAuth services from config.
func initAuthServices(ctx context.Context, v *viper.Viper, log *zap.Logger) (*authServices, error) {
	accessExpiry, _ := time.ParseDuration(v.GetString("jwt.access_expiry"))
	refreshExpiry, _ := time.ParseDuration(v.GetString("jwt.refresh_expiry"))

	jwtSvc := auth.NewJWTService(v.GetString("jwt.secret"), accessExpiry, refreshExpiry)
	passwordSvc := auth.NewPasswordService()

	svc := &authServices{JWT: jwtSvc, Password: passwordSvc}

	// OAuth service is optional — only initialized when client IDs are configured
	if gClientID := v.GetString("oauth.google.client_id"); gClientID != "" {
		oauthCfg := auth.OAuthConfig{
			GoogleClientID:        gClientID,
			GoogleClientSecret:    v.GetString("oauth.google.client_secret"),
			GoogleRedirectURL:     v.GetString("oauth.google.redirect_url"),
			MicrosoftClientID:     v.GetString("oauth.microsoft.client_id"),
			MicrosoftClientSecret: v.GetString("oauth.microsoft.client_secret"),
			MicrosoftTenantID:     v.GetString("oauth.microsoft.tenant_id"),
			MicrosoftRedirectURL:  v.GetString("oauth.microsoft.redirect_url"),
			FrontendCallbackURL:   v.GetString("oauth.frontend_callback_url"),
		}
		oauthSvc, err := auth.NewOAuthService(ctx, oauthCfg)
		if err != nil {
			log.Warn("oauth service init failed, continuing without OAuth", zap.Error(err))
		} else {
			svc.OAuth = oauthSvc
			log.Info("oauth service initialized")
		}
	} else {
		log.Info("oauth not configured (set oauth.google.client_id to enable)")
	}

	return svc, nil
}

// buildSelfURL returns core's own HTTP base URL for internal tool dispatch.
func buildSelfURL(v *viper.Viper) string {
	if u := v.GetString("server.self_url"); u != "" {
		return u
	}
	return fmt.Sprintf("http://localhost:%d", v.GetInt("server.http_port"))
}

// buildLLMProvider reads config and returns the configured LLM provider.
// Defaults to OpenAI-compatible if llm.provider is not set.
// Config keys:
//
//	llm.provider  = "openai" | "claude" | "gemini"
//	llm.api_key   = API key
//	llm.model     = model name
//	llm.base_url  = base URL (optional, for OpenAI-compat endpoints like Ollama)
func buildLLMProvider(v *viper.Viper) llm.LLMProvider {
	provider := v.GetString("llm.provider")
	apiKey := v.GetString("llm.api_key")
	model := v.GetString("llm.model")
	baseURL := v.GetString("llm.base_url")

	switch provider {
	case "mock":
		return llm.NewMockProvider()
	case "claude":
		if model == "" {
			model = "claude-haiku-4-5"
		}
		return llm.NewClaudeProvider(apiKey, model)
	case "gemini":
		if model == "" {
			model = "gemini-2.0-flash"
		}
		return llm.NewGeminiProvider(apiKey, model)
	default: // "openai" or any OpenAI-compatible
		if model == "" {
			model = "gpt-4o-mini"
		}
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return llm.NewOpenAIProvider(apiKey, model, baseURL)
	}
}
