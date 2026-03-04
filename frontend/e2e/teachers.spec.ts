import { test, expect } from './fixtures/app-fixture'

// Unique per test run to avoid code conflicts from previous runs
const runId = Date.now().toString(36)

test.describe('Teachers CRUD', () => {
  test('create and view a teacher', async ({ authedPage: page, adminToken }) => {
    await page.goto('/hr/teachers')
    await expect(page).toHaveURL(/hr\/teachers/)

    // Create a department first (required for teacher creation)
    // Uses admin token since POST /hr/departments requires admin/super_admin role
    const deptRes = await page.request.post('/api/hr/departments', {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { name: `E2E Teachers Dept ${runId}`, code: `TEACH${runId}` },
    })
    if (!deptRes.ok()) {
      throw new Error(`Dept creation failed: ${deptRes.status()} ${await deptRes.text()}`)
    }
    const dept = (await deptRes.json()) as { id: string }

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
