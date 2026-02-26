import { test, expect } from '@playwright/test'

const uniqueId = Date.now().toString(36)

test.describe('Authentication', () => {
  test('register a new user and redirect to login', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByRole('heading', { name: /create account/i })).toBeVisible()

    await page.getByLabel(/full name/i).fill('E2E Auth User')
    await page.getByLabel(/^email/i).fill(`auth-${uniqueId}@test.com`)
    await page.getByLabel(/^password/i).first().fill('Test123456')
    await page.getByLabel(/confirm password/i).fill('Test123456')
    await page.getByRole('button', { name: /create account/i }).click()

    // After registration, user should see success toast or be redirected to login
    await expect(page).toHaveURL(/login/, { timeout: 10_000 })
  })

  test('login with valid credentials and reach dashboard', async ({ page }) => {
    // First register via API
    const email = `login-${uniqueId}@test.com`
    const password = 'Test123456'
    await page.request.post('/api/auth/register', {
      data: { full_name: 'Login Test', email, password },
    })

    await page.goto('/login')
    await expect(page.getByRole('heading', { name: /myrmex erp/i })).toBeVisible()

    await page.getByLabel(/email/i).fill(email)
    await page.getByLabel(/password/i).fill(password)
    await page.getByRole('button', { name: /sign in/i }).click()

    await expect(page).toHaveURL(/dashboard/, { timeout: 10_000 })
  })

  test('login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/login')
    await page.getByLabel(/email/i).fill('invalid@test.com')
    await page.getByLabel(/password/i).fill('wrongpassword')
    await page.getByRole('button', { name: /sign in/i }).click()

    // Should show error toast
    await expect(page.getByText(/login failed|invalid/i)).toBeVisible({ timeout: 5_000 })
  })
})
