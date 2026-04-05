# 2026-04-05 — Fix OAuth Login Across All Environments

## Summary

Fixed OAuth login failures in Docker Compose, Terraform (prod/staging), and Cloud Run. Found and patched 4 distinct root causes spanning environment config, TLS termination, and URL encoding. All 87 tests pass, build clean.

## Root Causes & Fixes

### 1. Docker Compose Missing OAuth Env Vars
**Symptom**: Login endpoints returned 404 in local dev. OAuth routes never registered.

**Root Cause**: `docker-compose.yml` core service had no `OAUTH_*` environment variables. Service startup skipped OAuth initialization entirely.

**Fix**: Added 5 required env vars to compose service:
- `OAUTH_ENABLED=true`
- `OAUTH_GOOGLE_CLIENT_ID`, `OAUTH_GOOGLE_CLIENT_SECRET`
- `OAUTH_MICROSOFT_CLIENT_ID`, `OAUTH_MICROSOFT_CLIENT_SECRET`

Impact: OAuth routes now initialize in dev environment.

### 2. Terraform Missing Redirect URL Env Vars
**Symptom**: Staging/production OAuth secrets existed but login redirects failed. OAuth provider validation rejected mismatched redirect URLs.

**Root Cause**: `terraform/main.tf` set client ID/secret secrets but omitted `OAUTH_GOOGLE_REDIRECT_URL`, `OAUTH_MICROSOFT_REDIRECT_URL`, `OAUTH_FRONTEND_CALLBACK_URL`. Service fell back to hardcoded localhost defaults.

**Fix**: Added redirect URL env vars derived from domain variables. Added `staging_api_domain` and `staging_frontend_domain` variables to Terraform, templated URLs:
```
OAUTH_GOOGLE_REDIRECT_URL = "https://${var.api_domain}/auth/callback/google"
OAUTH_FRONTEND_CALLBACK_URL = "https://${var.frontend_domain}/login/callback"
```

Impact: OAuth provider redirect URLs now match actual deployment domains.

### 3. Cookie Secure Flag Broken Behind Proxy
**Symptom**: Browsers rejected auth cookies on HTTPS. Cookie flagged as non-Secure despite HTTPS connection.

**Root Cause**: TLS termination at Cloud Run load balancer or nginx proxy. `c.Request.TLS != nil` always false, service thought connection was HTTP → set `Secure=false` on cookies.

**Fix**: Extended secure cookie detection to check `X-Forwarded-Proto` header:
```go
isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
```

Impact: Cookies now marked Secure in proxied environments. Browsers accept on HTTPS.

### 4. URL Encoding Bug in Error Redirects
**Symptom**: Error paths produced malformed redirect URLs.

**Root Cause**: Code review found `FrontendCallbackURL` concatenation didn't URL-encode the authorization code parameter. Codes containing `&` or `=` broke redirect string.

**Fix**: Wrapped code in `url.QueryEscape()`:
```go
redirect := fmt.Sprintf("%s?code=%s", frontend_callback_url, url.QueryEscape(code))
```

Impact: Error redirects now produce valid URLs.

## Also Completed

- Created `docs/oauth-provider-setup.md` (Google + Microsoft setup walkthrough)
- Updated `deployment-guide.md` with environment variable checklist
- Updated `project-roadmap.md` phase 4 completion status
- Ran full test suite: **87 tests pass**, build clean

## Code Review Findings (Non-Blocking)

- **Scaling blocker**: In-memory auth code store shared across goroutines. Multiple Cloud Run instances need Redis for distributed state.
- **Terraform incomplete**: Staging domain mapping resources missing if custom domains required.
- **Defense-in-depth**: Consider `gin.SetTrustedProxies()` for explicit proxy IP allowlist.

All flagged for future work, don't block this merge.

## Commit

`c00eef6 fix(oauth): resolve Docker Compose env vars, Terraform redirect URLs, cookie Secure flag, and URL encoding`

## Lessons

1. **Environment isolation breaks on config variance** — Docker Compose, Terraform, and Cloud Run use different env var sources. Test each deployment path with live OAuth provider integration before merge.

2. **TLS termination makes debugging painful** — The real connection is encrypted, request object lies about it. Header inspection is mandatory for proxy scenarios. Document this expectation for future developers.

3. **URL parameter handling is easy to miss** — Authorization codes should be treated as untrusted data. Always escape before serializing to URLs. Add linting rule or unit test asserting URL validity.

4. **OAuth provider mismatch is silent** — Google/Microsoft don't 404 on wrong redirect URL; they reject the entire flow with vague error messages. Validate redirect URLs match actual deployment before deployment.

## Next Steps

1. Monitor production login success rate post-deploy. Watch for any remaining OAuth failures in logs.
2. Plan Redis integration for distributed auth code store (phase 5 or later).
3. Add integration test that validates OAuth flow with mock provider across Docker/Terraform/Cloud Run configs.
