import { authStore } from './auth-store'

// jsdom provides localStorage
beforeEach(() => {
  localStorage.clear()
})

describe('authStore', () => {
  it('stores and retrieves tokens', () => {
    authStore.setTokens('access-123', 'refresh-456')
    expect(authStore.getAccessToken()).toBe('access-123')
    expect(authStore.getRefreshToken()).toBe('refresh-456')
  })

  it('stores and retrieves user', () => {
    const user = { id: '1', email: 'test@test.com', username: 'test', role: 'user' }
    authStore.setUser(user as never)
    const stored = authStore.getUser()
    expect(stored?.email).toBe('test@test.com')
  })

  it('returns null for missing user', () => {
    expect(authStore.getUser()).toBeNull()
  })

  it('returns null for corrupted user JSON', () => {
    localStorage.setItem('auth_user', '{invalid')
    expect(authStore.getUser()).toBeNull()
  })

  it('reports authentication status', () => {
    expect(authStore.isAuthenticated()).toBe(false)
    authStore.setTokens('tok', 'ref')
    expect(authStore.isAuthenticated()).toBe(true)
  })

  it('clears all auth data', () => {
    authStore.setTokens('tok', 'ref')
    authStore.clear()
    expect(authStore.getAccessToken()).toBeNull()
    expect(authStore.getRefreshToken()).toBeNull()
    expect(authStore.getUser()).toBeNull()
  })
})
