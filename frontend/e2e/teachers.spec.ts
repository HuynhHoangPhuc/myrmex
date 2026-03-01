import { test, expect } from './fixtures/app-fixture'

test.describe('Teachers CRUD', () => {
  test('create and view a teacher', async ({ authedPage: page }) => {
    await page.goto('/hr/teachers')
    await expect(page).toHaveURL(/hr\/teachers/)

    // Create a department first (required for teacher creation)
    const token = await page.evaluate(() => localStorage.getItem('access_token'))
    const deptRes = await page.request.post('/api/hr/departments', {
      data: { name: 'E2E Teachers Dept', code: 'E2ETEACH' },
      headers: { Authorization: `Bearer ${token}` },
    })
    const dept = await deptRes.json()

    await page.getByRole('link', { name: /new|add|create/i }).click()
    await expect(page).toHaveURL(/hr\/teachers\/new/)

    await page.getByLabel(/employee code/i).fill('E2E-001')
    await page.getByLabel(/full name/i).fill('E2E Test Teacher')
    await page.getByLabel(/email/i).fill('e2e-teacher@test.com')

    // Wait for department option to load, then select it
    await expect(page.locator(`option[value="${dept.id}"]`)).toBeAttached({ timeout: 5_000 })
    await page.locator('select').selectOption(dept.id)

    await page.getByRole('button', { name: /create|save|submit/i }).click()

    // Navigates to teacher detail page (URL leaves /new)
    await page.waitForURL(
      (url) => url.pathname.includes('/hr/teachers/') && !url.pathname.endsWith('/new'),
      { timeout: 10_000 },
    )
    await expect(page.getByText('E2E Test Teacher')).toBeVisible({ timeout: 5_000 })
  })
})
