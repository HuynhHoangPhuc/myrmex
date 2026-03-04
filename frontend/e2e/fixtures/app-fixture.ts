import { test as base, expect, type Page } from '@playwright/test'

// Seeded admin credentials (from deploy/docker/seed.sql)
const ADMIN_EMAIL = 'admin@myrmex.dev'
const ADMIN_PASSWORD = 'demo1234'

interface AuthFixtures {
  /** Pre-authenticated page (logged in as seeded admin) */
  authedPage: Page
  /** Admin bearer token for setup API calls requiring elevated permissions */
  adminToken: string
}

export const test = base.extend<AuthFixtures>({
  adminToken: async ({ request }, use) => {
    const loginRes = await request.post('/api/auth/login', {
      data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    })
    if (!loginRes.ok()) {
      throw new Error(`Admin login failed: ${loginRes.status()} ${await loginRes.text()}`)
    }
    const { access_token } = await loginRes.json()
    await use(access_token as string)
  },

  authedPage: async ({ page }, use) => {
    // Use seeded admin account so HR/department routes are accessible in tests
    // (GET /api/hr/departments requires admin/dept_head role)
    const email = ADMIN_EMAIL
    const password = ADMIN_PASSWORD

    // Login
    const loginRes = await page.request.post('/api/auth/login', {
      data: { email, password },
    })
    expect(loginRes.ok()).toBeTruthy()

    const { access_token, refresh_token, user } = await loginRes.json()

    // Inject tokens into localStorage so the SPA picks them up
    await page.addInitScript(
      ({ accessToken, refreshToken, userData }) => {
        localStorage.setItem('access_token', accessToken)
        localStorage.setItem('refresh_token', refreshToken)
        localStorage.setItem('auth_user', JSON.stringify(userData))
      },
      { accessToken: access_token, refreshToken: refresh_token, userData: user },
    )

    await use(page)
  },
})

export { expect }
