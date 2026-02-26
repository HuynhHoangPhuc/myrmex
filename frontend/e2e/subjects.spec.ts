import { test, expect } from './fixtures/app-fixture'

test.describe('Subjects CRUD', () => {
  test('create and view a subject', async ({ authedPage: page }) => {
    await page.goto('/subjects')
    await expect(page).toHaveURL(/subjects/)

    await page.getByRole('link', { name: /new|add|create/i }).click()
    await expect(page).toHaveURL(/subjects\/new/)

    await page.getByLabel(/code/i).fill('E2E101')
    await page.getByLabel(/name/i).first().fill('E2E Test Subject')
    await page.getByLabel(/credits/i).fill('3')

    await page.getByRole('button', { name: /create|save|submit/i }).click()

    await page.waitForURL(/subjects/, { timeout: 10_000 })
    await expect(page.getByText('E2E Test Subject')).toBeVisible({ timeout: 5_000 })
  })
})
