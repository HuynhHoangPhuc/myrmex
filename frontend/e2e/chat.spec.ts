import { test, expect } from './fixtures/app-fixture'

test.describe('AI Chat (Mock Provider)', () => {
  test('open chat panel and send a message', async ({ authedPage: page }) => {
    await page.goto('/dashboard')

    // Wait for toggle button then open chat panel
    const chatToggle = page.getByRole('button', { name: /chat|ai|assistant/i }).first()
    await chatToggle.waitFor({ state: 'visible', timeout: 10_000 })
    await chatToggle.click()

    // Find chat input and send a message
    const chatInput = page.getByPlaceholder(/message|ask|type/i)
    await expect(chatInput).toBeVisible({ timeout: 10_000 })
    await chatInput.fill('hello')
    await chatInput.press('Enter')

    // Mock provider should return "This is a mock response for: hello"
    await expect(page.getByText(/mock response/i)).toBeVisible({ timeout: 10_000 })
  })

  test('trigger a tool call via chat', async ({ authedPage: page }) => {
    await page.goto('/dashboard')

    const chatToggle = page.getByRole('button', { name: /chat|ai|assistant/i }).first()
    await chatToggle.waitFor({ state: 'visible', timeout: 10_000 })
    await chatToggle.click()

    const chatInput = page.getByPlaceholder(/message|ask|type/i)
    await expect(chatInput).toBeVisible({ timeout: 10_000 })
    await chatInput.fill('list teachers')
    await chatInput.press('Enter')

    // Should show some tool execution feedback (tool call or result)
    await expect(
      page.getByText(/list_teachers|teacher|executing/i).first(),
    ).toBeVisible({ timeout: 10_000 })
  })
})
