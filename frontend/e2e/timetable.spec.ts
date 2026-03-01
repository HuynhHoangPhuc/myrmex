import { test, expect } from './fixtures/app-fixture'

test.describe('Timetable', () => {
  test('view semesters page', async ({ authedPage: page }) => {
    await page.goto('/timetable/semesters')
    await expect(page).toHaveURL(/timetable\/semesters/)
    // Page should load without errors
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 5_000 })
  })

  test('create a semester', async ({ authedPage: page }) => {
    await page.goto('/timetable/semesters/new')
    await expect(page).toHaveURL(/timetable\/semesters\/new/)

    // Fill all required fields (name, start date, end date; year/term have defaults)
    await page.getByLabel(/name/i).fill('E2E Semester 2026')
    await page.getByLabel(/start date/i).fill('2026-09-01')
    await page.getByLabel(/end date/i).fill('2026-12-31')

    await page.getByRole('button', { name: /create semester/i }).click()

    // Wizard moves to step 2 after creation — URL gains semesterId param
    await page.waitForURL((url) => url.href.includes('semesterId'), { timeout: 10_000 })

    // Navigate to semesters list to verify the semester appears
    await page.goto('/timetable/semesters')
    await expect(page.getByText('E2E Semester 2026')).toBeVisible({ timeout: 5_000 })
  })
})
