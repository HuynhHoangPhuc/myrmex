import axios from 'axios'
import { ENDPOINTS } from './endpoints'

// Axios instance with base URL from env, 30s timeout
export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  timeout: 30_000,
})

// Attach JWT token to every request
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Track whether a token refresh is in progress to avoid concurrent refreshes
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

function redirectToLogin() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('auth_user')
  const isAuthPage = ['/login', '/register'].some((p) =>
    window.location.pathname.startsWith(p),
  )
  if (!isAuthPage) {
    window.location.href = '/login'
  }
}

// On 401: attempt silent token refresh, retry original request; redirect to login only if refresh fails
apiClient.interceptors.response.use(
  (res) => res,
  async (err: unknown) => {
    const axiosErr = err as { response?: { status?: number }; config?: { _retry?: boolean; headers?: Record<string, string>; url?: string } }
    const status = axiosErr?.response?.status
    const config = axiosErr?.config

    // Only attempt refresh for 401s on non-refresh endpoints that haven't already been retried
    if (status !== 401 || config?._retry || config?.url === ENDPOINTS.auth.refresh) {
      return Promise.reject(err)
    }

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      redirectToLogin()
      return Promise.reject(err)
    }

    if (isRefreshing) {
      // Queue this request until the ongoing refresh completes
      return new Promise((resolve, reject) => {
        pendingRequests.push((newToken) => {
          if (!config) return reject(err)
          config._retry = true
          config.headers = { ...config.headers, Authorization: `Bearer ${newToken}` }
          resolve(apiClient(config))
        })
      })
    }

    isRefreshing = true
    try {
      const { data } = await apiClient.post<{ access_token: string; refresh_token: string }>(
        ENDPOINTS.auth.refresh,
        { refresh_token: refreshToken },
      )
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)

      // Flush queued requests with the new token
      pendingRequests.forEach((cb) => cb(data.access_token))
      pendingRequests = []

      // Retry the original request
      if (!config) return Promise.reject(err)
      config._retry = true
      config.headers = { ...config.headers, Authorization: `Bearer ${data.access_token}` }
      return apiClient(config)
    } catch {
      pendingRequests = []
      redirectToLogin()
      return Promise.reject(err)
    } finally {
      isRefreshing = false
    }
  },
)
