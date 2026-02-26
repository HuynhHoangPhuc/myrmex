import { test, expect } from './fixtures/app-fixture'

test.describe('Teachers CRUD', () => {
  test('create and view a teacher', async ({ authedPage: page }) => {
    // Navigate to HR teachers page
    await page.goto('/hr/teachers')
    await expect(page).toHaveURL(/hr\/teachers/)

    // Click "New Teacher" or similar button
    await page.getByRole('link', { name: /new|add|create/i }).click()
    await expect(page).toHaveURL(/hr\/teachers\/new/)

    // Fill teacher form
    await page.getByLabel(/employee code/i).fill('E2E-001')
    await page.getByLabel(/full name/i).fill('E2E Test Teacher')
    await page.getByLabel(/email/i).fill('e2e-teacher@test.com')

    // Submit form
    await page.getByRole('button', { name: /create|save|submit/i }).click()

    // Should redirect back to list or detail page
    await page.waitForURL(/hr\/teachers/, { timeout: 10_000 })

    // Teacher should appear in the list
    await expect(page.getByText('E2E Test Teacher')).toBeVisible({ timeout: 5_000 })
  })
})
