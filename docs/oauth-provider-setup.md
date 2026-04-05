# OAuth Provider Setup Guide

Setup guide for Google and Microsoft OAuth/SSO integration with Myrmex.

## Prerequisites

- Google Cloud project with billing enabled
- Azure AD tenant (HCMUS institutional)
- Domain names decided for production/staging

## 1. Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com) → **APIs & Services** → **Credentials**
2. Configure **OAuth consent screen** (if not done):
   - User type: Internal (restrict to `hcmus.edu.vn`)
   - App name: Myrmex
   - Scopes: `openid`, `email`, `profile`
3. Create **OAuth 2.0 Client ID** → Web application
4. Add **Authorized redirect URIs**:

| Environment | Redirect URI |
|-------------|-------------|
| Local | `http://localhost:8080/api/auth/oauth/google/callback` |
| Staging | `https://<staging-api-domain>/api/auth/oauth/google/callback` |
| Production | `https://<api-domain>/api/auth/oauth/google/callback` |

5. Copy **Client ID** and **Client Secret**

## 2. Microsoft (Azure AD) Setup

1. Go to [Azure Portal](https://portal.azure.com) → **Azure Active Directory** → **App registrations**
2. **New registration**:
   - Name: Myrmex
   - Supported account types: Single tenant (HCMUS)
   - Platform: Web
3. Add **Redirect URIs**:

| Environment | Redirect URI |
|-------------|-------------|
| Local | `http://localhost:8080/api/auth/oauth/microsoft/callback` |
| Staging | `https://<staging-api-domain>/api/auth/oauth/microsoft/callback` |
| Production | `https://<api-domain>/api/auth/oauth/microsoft/callback` |

4. Note from **Overview**: Application (client) ID, Directory (tenant) ID
5. Go to **Certificates & secrets** → New client secret → copy value

## 3. Local Development (Docker Compose)

Create `deploy/docker/.env` (gitignored):

```bash
OAUTH_GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
OAUTH_GOOGLE_CLIENT_SECRET=GOCSPX-xxxxx
OAUTH_MICROSOFT_CLIENT_ID=xxxxx-xxxxx-xxxxx
OAUTH_MICROSOFT_CLIENT_SECRET=xxxxx~xxxxx
OAUTH_MICROSOFT_TENANT_ID=xxxxx-xxxxx-xxxxx
```

Run `make demo` — OAuth buttons should work. Without `.env`, demo boots normally with password-only login.

## 4. Production / Staging (GCP)

Credentials are stored in **GCP Secret Manager** (managed by Terraform). Redirect URLs are derived from domain variables.

Set in `deploy/terraform/terraform.tfvars` (gitignored):

```hcl
api_domain              = "api.myrmex.hcmus.edu.vn"
frontend_domain         = "myrmex.hcmus.edu.vn"
staging_api_domain      = "staging-api.myrmex.hcmus.edu.vn"
staging_frontend_domain = "staging.myrmex.hcmus.edu.vn"
```

Run `terraform apply` — deploys with correct redirect URLs.

## 5. Email Domain Restrictions

| Email Domain | Assigned Role | Provider |
|---|---|---|
| `@hcmus.edu.vn` | teacher | Google |
| `@student.hcmus.edu.vn` | student | Microsoft |

Users with other domains are rejected. Accounts must be pre-created by admin before first OAuth login.

## 6. Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `redirect_uri_mismatch` | Redirect URI in provider console doesn't match config | Verify URIs match exactly (scheme, host, path) |
| `404` on `/api/auth/oauth/*/login` | OAuth credentials not configured | Check env vars are set and non-empty |
| `missing oauth state cookie` | Cookie blocked or domain mismatch | Check `Secure` flag, `SameSite`, cross-domain issues |
| `unauthorized email domain` | Email not `@hcmus.edu.vn` or `@student.hcmus.edu.vn` | Use institutional email |
| `no account found` | No pre-created teacher/student record | Admin must create account first |
