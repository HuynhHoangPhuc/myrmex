import { test, expect } from './fixtures/app-fixture'

// Unique per test run to avoid code conflicts from previous runs
const runId = Date.now().toString(36)

test.describe('Subjects CRUD', () => {
  test('create and view a subject', async ({ authedPage: page, adminToken }) => {
    await page.goto('/subjects')
    await expect(page).toHaveURL(/subjects/)

    // Create a department first (required for subject creation)
    // Uses admin token since POST /hr/departments requires admin/super_admin role
    const deptRes = await page.request.post('/api/hr/departments', {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { name: `E2E Subject Dept ${runId}`, code: `SUBJ${runId}` },
    })
    if (!deptRes.ok()) {
      throw new Error(`Dept creation failed: ${deptRes.status()} ${await deptRes.text()}`)
    }
    const dept = (await deptRes.json()) as { id: string }

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
