import { test as base, expect, type Page } from '@playwright/test'

// Unique user per test run to avoid conflicts
const testId = Date.now().toString(36)

// Seeded admin credentials (from deploy/docker/seed.sql)
const ADMIN_EMAIL = 'admin@myrmex.dev'
const ADMIN_PASSWORD = 'demo1234'

interface AuthFixtures {
  /** Pre-authenticated page (registered + logged in as viewer) */
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
    const email = `e2e-${testId}@test.com`
    const password = 'Test123456'

    // Register
    const regRes = await page.request.post('/api/auth/register', {
      data: { full_name: 'E2E Test User', email, password },
    })
    // If already registered (409) or rate-limited (429), continue to login
    if (!regRes.ok() && regRes.status() !== 409 && regRes.status() !== 429) {
      throw new Error(`Registration failed: ${regRes.status()} ${await regRes.text()}`)
    }

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
