import { test, expect } from './fixtures/app-fixture'

// Unique per test run to avoid code conflicts from previous runs
const runId = Date.now().toString(36)

test.describe('Teachers CRUD', () => {
  test('create and view a teacher', async ({ authedPage: page }) => {
    await page.goto('/hr/teachers')
    await expect(page).toHaveURL(/hr\/teachers/)

    // Create a department first (required for teacher creation)
    // Use page.evaluate so the fetch runs in the browser context with the correct token
    const dept = await page.evaluate(async (id) => {
      const token = localStorage.getItem('access_token')
      const res = await fetch('/api/hr/departments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: `E2E Teachers Dept ${id}`, code: `TEACH${id}` }),
      })
      if (!res.ok) throw new Error(`Dept creation failed: ${res.status} ${await res.text()}`)
      return (await res.json()) as { id: string }
    }, runId)

    await page.getByRole('link', { name: /new|add|create/i }).click()
    await expect(page).toHaveURL(/hr\/teachers\/new/)

    await page.getByLabel(/employee code/i).fill(`E2E-${runId}`)
    await page.getByLabel(/full name/i).fill('E2E Test Teacher')
    await page.getByLabel(/email/i).fill(`e2e-teacher-${runId}@test.com`)
    await page.getByLabel(/title/i).fill('Dr.')

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
