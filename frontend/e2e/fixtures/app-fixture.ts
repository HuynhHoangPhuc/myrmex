import { test as base, expect, type Page } from '@playwright/test'

// Unique user per test run to avoid conflicts
const testId = Date.now().toString(36)

interface AuthFixtures {
  /** Pre-authenticated page (registered + logged in) */
  authedPage: Page
}

export const test = base.extend<AuthFixtures>({
  authedPage: async ({ page }, use) => {
    const email = `e2e-${testId}@test.com`
    const password = 'Test123456'

    // Register
    const regRes = await page.request.post('/api/auth/register', {
      data: { full_name: 'E2E Test User', email, password },
    })
    // If already registered (409), continue to login
    if (!regRes.ok() && regRes.status() !== 409) {
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
