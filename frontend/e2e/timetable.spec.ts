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

    await page.getByLabel(/name/i).fill('E2E Semester 2026')

    await page.getByRole('button', { name: /create|save|submit/i }).click()

    await page.waitForURL(/timetable\/semesters/, { timeout: 10_000 })
    await expect(page.getByText('E2E Semester 2026')).toBeVisible({ timeout: 5_000 })
  })
})
