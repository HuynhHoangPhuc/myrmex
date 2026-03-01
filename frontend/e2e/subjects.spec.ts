import { test, expect } from './fixtures/app-fixture'

test.describe('Subjects CRUD', () => {
  test('create and view a subject', async ({ authedPage: page }) => {
    await page.goto('/subjects')
    await expect(page).toHaveURL(/subjects/)

    // Create a department first (required for subject creation)
    const token = await page.evaluate(() => localStorage.getItem('access_token'))
    const deptRes = await page.request.post('/api/hr/departments', {
      data: { name: 'E2E Subject Dept', code: 'E2ESUBJ' },
      headers: { Authorization: `Bearer ${token}` },
    })
    const dept = await deptRes.json()

    await page.getByRole('link', { name: /new|add|create/i }).click()
    await expect(page).toHaveURL(/subjects\/new/)

    await page.getByLabel(/code/i).fill('E2E101')
    await page.getByLabel(/name/i).first().fill('E2E Test Subject')
    await page.getByLabel(/credits/i).fill('3')

    // Wait for department option to load, then select it
    await expect(page.locator(`option[value="${dept.id}"]`)).toBeAttached({ timeout: 5_000 })
    await page.locator('select').selectOption(dept.id)

    await page.getByRole('button', { name: /create|save|submit/i }).click()

    // Navigates to subject detail page (URL leaves /new)
    await page.waitForURL(
      (url) => url.pathname.includes('/subjects/') && !url.pathname.endsWith('/new'),
      { timeout: 10_000 },
    )
    await expect(page.getByText('E2E Test Subject')).toBeVisible({ timeout: 5_000 })
  })
})
