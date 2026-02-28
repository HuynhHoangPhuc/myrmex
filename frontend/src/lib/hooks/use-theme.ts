import * as React from 'react'

type Theme = 'light' | 'dark'

const THEME_STORAGE_KEY = 'theme'

/** Reads current theme from the DOM class set by main.tsx at startup. */
function getInitialTheme(): Theme {
  if (typeof document !== 'undefined') {
    return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
  }
  return 'light'
}

export function useTheme() {
  const [theme, setThemeState] = React.useState<Theme>(getInitialTheme)

  // Sync DOM class whenever theme state changes
  React.useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, [theme])

  // When no explicit preference is stored, follow OS changes in real-time
  React.useEffect(() => {
    if (typeof window === 'undefined') return
    if (window.localStorage.getItem(THEME_STORAGE_KEY)) return // user has explicit choice

    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (e: MediaQueryListEvent) => {
      setThemeState(e.matches ? 'dark' : 'light')
    }
    mq.addEventListener('change', handleChange)
    return () => mq.removeEventListener('change', handleChange)
  }, [])

  /** Explicitly set theme — saves to localStorage so OS changes are no longer followed. */
  const setTheme = React.useCallback((newTheme: Theme) => {
    window.localStorage.setItem(THEME_STORAGE_KEY, newTheme)
    setThemeState(newTheme)
  }, [])

  const toggle = React.useCallback(() => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }, [theme, setTheme])

  return { theme, setTheme, toggle }
}
