import * as React from 'react'

type Theme = 'light' | 'dark'

const THEME_STORAGE_KEY = 'theme'

function getInitialTheme(): Theme {
  if (typeof document !== 'undefined') {
    return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
  }

  return 'light'
}

function applyTheme(theme: Theme) {
  if (typeof document === 'undefined') return

  document.documentElement.classList.toggle('dark', theme === 'dark')
  window.localStorage.setItem(THEME_STORAGE_KEY, theme)
}

export function useTheme() {
  const [theme, setTheme] = React.useState<Theme>(() => getInitialTheme())

  React.useEffect(() => {
    applyTheme(theme)
  }, [theme])

  const toggle = React.useCallback(() => {
    setTheme((current) => (current === 'dark' ? 'light' : 'dark'))
  }, [])

  return { theme, setTheme, toggle }
}
