import { createRootRouteWithContext, Outlet } from '@tanstack/react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import type { QueryClient } from '@tanstack/react-query'
import { queryClient } from '@/config/query-client'

interface RouterContext {
  queryClient: QueryClient
}

// Root layout: wraps entire app with QueryClientProvider
// ReactQueryDevtools only in dev mode
export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => (
    <QueryClientProvider client={queryClient}>
      <Outlet />
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  ),
})
