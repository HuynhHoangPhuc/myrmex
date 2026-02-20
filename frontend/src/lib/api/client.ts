import axios from 'axios'

// Axios instance with base URL from env, 30s timeout
export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
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

// On 401 clear tokens and redirect to login
apiClient.interceptors.response.use(
  (res) => res,
  (err: unknown) => {
    const status = (err as { response?: { status?: number } })?.response?.status
    if (status === 401) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  },
)
