import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { NAV_ITEMS } from '@/components/layouts/sidebar-nav'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { ListResponse } from '@/lib/api/types'
import type { Teacher } from '@/modules/hr/types'
import type { Subject } from '@/modules/subject/types'

interface CommandPaletteProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

interface CommandItem {
  id: string
  label: string
  hint: string
  path: string
}

const STATIC_PAGES: CommandItem[] = NAV_ITEMS.flatMap((item) => {
  const pages: CommandItem[] = [
    {
      id: item.to,
      label: item.label,
      hint: 'Page',
      path: item.to,
    },
  ]

  item.children?.forEach((child) => {
    pages.push({
      id: child.to,
      label: child.label,
      hint: item.label,
      path: child.to,
    })
  })

  return pages
})

function filterItems(items: CommandItem[], query: string): CommandItem[] {
  const normalized = query.trim().toLowerCase()
  if (!normalized) return items.slice(0, 8)

  return items.filter((item) => item.label.toLowerCase().includes(normalized)).slice(0, 8)
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const [search, setSearch] = React.useState('')

  const { data: teachers = [] } = useQuery({
    queryKey: ['command-palette', 'teachers'],
    queryFn: async () => {
      const { data } = await apiClient.get<ListResponse<Teacher>>(ENDPOINTS.hr.teachers, {
        params: { page: 1, page_size: 20 },
      })
      return data.data
    },
    enabled: open,
    staleTime: 60_000,
  })

  const { data: subjects = [] } = useQuery({
    queryKey: ['command-palette', 'subjects'],
    queryFn: async () => {
      const { data } = await apiClient.get<ListResponse<Subject>>(ENDPOINTS.subjects.list, {
        params: { page: 1, page_size: 50 },
      })
      return data.data
    },
    enabled: open,
    staleTime: 60_000,
  })

  React.useEffect(() => {
    if (!open) setSearch('')
  }, [open])

  const pages = React.useMemo(() => filterItems(STATIC_PAGES, search), [search])
  const teacherItems = React.useMemo(
    () =>
      filterItems(
        teachers.map((teacher) => ({
          id: teacher.id,
          label: teacher.full_name,
          hint: 'Teacher',
          path: `/hr/teachers/${teacher.id}`,
        })),
        search,
      ),
    [search, teachers],
  )
  const subjectItems = React.useMemo(
    () =>
      filterItems(
        subjects.map((subject) => ({
          id: subject.id,
          label: `${subject.code} · ${subject.name}`,
          hint: 'Subject',
          path: `/subjects/${subject.id}`,
        })),
        search,
      ),
    [search, subjects],
  )

  const hasResults = pages.length > 0 || teacherItems.length > 0 || subjectItems.length > 0

  function handleSelect(path: string) {
    onOpenChange(false)
    window.location.href = path
  }

  function renderSection(title: string, items: CommandItem[]) {
    if (items.length === 0) return null

    return (
      <div className="space-y-1">
        <p className="px-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</p>
        <div className="space-y-1">
          {items.map((item) => (
            <button
              key={item.id}
              type="button"
              className="flex w-full items-center justify-between rounded-md px-2 py-2 text-left text-sm transition-colors hover:bg-muted"
              onClick={() => handleSelect(item.path)}
            >
              <span className="truncate">{item.label}</span>
              <span className="ml-3 shrink-0 text-xs text-muted-foreground">{item.hint}</span>
            </button>
          ))}
        </div>
      </div>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Command palette</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search pages, teachers, subjects..."
              className="pl-9"
              autoFocus
            />
          </div>

          <div className="max-h-[420px] space-y-4 overflow-y-auto pr-1">
            {hasResults ? (
              <>
                {renderSection('Pages', pages)}
                {renderSection('Teachers', teacherItems)}
                {renderSection('Subjects', subjectItems)}
              </>
            ) : (
              <p className="py-8 text-center text-sm text-muted-foreground">No matches found.</p>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
