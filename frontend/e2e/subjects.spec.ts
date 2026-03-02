import { test, expect } from './fixtures/app-fixture'

// Unique per test run to avoid code conflicts from previous runs
const runId = Date.now().toString(36)

test.describe('Subjects CRUD', () => {
  test('create and view a subject', async ({ authedPage: page }) => {
    await page.goto('/subjects')
    await expect(page).toHaveURL(/subjects/)

    // Create a department first (required for subject creation)
    // Use page.evaluate so the fetch runs in the browser context with the correct token
    const dept = await page.evaluate(async (id) => {
      const token = localStorage.getItem('access_token')
      const res = await fetch('/api/hr/departments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: `E2E Subject Dept ${id}`, code: `SUBJ${id}` }),
      })
      if (!res.ok) throw new Error(`Dept creation failed: ${res.status} ${await res.text()}`)
      return (await res.json()) as { id: string }
    }, runId)

    await page.getByRole('link', { name: /new|add|create/i }).click()
    await expect(page).toHaveURL(/subjects\/new/)

    await page.getByLabel(/code/i).fill(`E2E${runId}`)
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
    // Use heading to avoid strict mode violation when subject name appears multiple times
    await expect(page.getByRole('heading', { name: 'E2E Test Subject' })).toBeVisible({
      timeout: 5_000,
    })
  })
})
