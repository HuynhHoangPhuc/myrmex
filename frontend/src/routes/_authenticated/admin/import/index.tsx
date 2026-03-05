import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { PageHeader } from '@/components/shared/page-header'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ImportPreviewTable } from './-import-preview-table'
import { ImportResults } from './-import-results'
import type { ImportResult } from './-import-results'

export const Route = createFileRoute('/_authenticated/admin/import/')({
  component: BulkImportPage,
})

type ImportType = 'teachers' | 'students'

const TABS: { id: ImportType; label: string }[] = [
  { id: 'teachers', label: 'Teachers' },
  { id: 'students', label: 'Students' },
]

function downloadTemplate(type: ImportType) {
  apiClient
    .get(`/admin/import/template/${type}`, { responseType: 'blob' })
    .then((res) => {
      const url = URL.createObjectURL(res.data as Blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${type}-template.csv`
      a.click()
      URL.revokeObjectURL(url)
    })
    .catch(() => {
      // Silent: browser will show nothing; real errors surface via network tab
    })
}

function useImport(type: ImportType) {
  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await apiClient.post<ImportResult>(
        `/admin/import/${type}`,
        formData,
        { headers: { 'Content-Type': 'multipart/form-data' } },
      )
      return data
    },
  })
}

interface TabPanelState {
  file: File | null
  csvText: string
  result: ImportResult | null
}

const EMPTY_STATE: TabPanelState = { file: null, csvText: '', result: null }

function BulkImportPage() {
  const [activeTab, setActiveTab] = React.useState<ImportType>('teachers')
  // Track state per-tab so switching tabs preserves each tab's file + result
  const [panelState, setPanelState] = React.useState<Record<ImportType, TabPanelState>>({
    teachers: { ...EMPTY_STATE },
    students: { ...EMPTY_STATE },
  })

  const mutation = useImport(activeTab)
  const current = panelState[activeTab]

  function updateCurrent(patch: Partial<TabPanelState>) {
    setPanelState((prev) => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], ...patch },
    }))
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0] ?? null
    if (!file) {
      updateCurrent({ file: null, csvText: '', result: null })
      return
    }

    mutation.reset()
    const reader = new FileReader()
    reader.onload = (ev) => {
      updateCurrent({ file, csvText: (ev.target?.result as string) ?? '', result: null })
    }
    reader.readAsText(file)
    // Reset the input so re-selecting the same file triggers onChange again
    e.target.value = ''
  }

  async function handleImport() {
    if (!current.file) return
    try {
      const result = await mutation.mutateAsync(current.file)
      updateCurrent({ result })
    } catch {
      // Error surfaced via mutation.error
    }
  }

  function handleTabChange(tab: ImportType) {
    setActiveTab(tab)
    mutation.reset()
  }

  const errorMessage = mutation.error
    ? ((mutation.error as { message?: string }).message ?? 'Import failed')
    : null

  return (
    <div className="space-y-6">
      <PageHeader
        title="Bulk Import"
        description="Import teachers or students from a CSV file"
      />

      {/* Tab bar */}
      <div className="flex border-b">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => handleTabChange(tab.id)}
            className={[
              'px-5 py-2.5 text-sm font-medium border-b-2 transition-colors',
              activeTab === tab.id
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground',
            ].join(' ')}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <Card className="p-6 space-y-5">
        {/* Template download */}
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">Step 1 — Download template</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Fill in the CSV template, then upload it below.
            </p>
          </div>
          <Button variant="outline" size="sm" onClick={() => downloadTemplate(activeTab)}>
            Download Template
          </Button>
        </div>

        <hr className="border-border" />

        {/* File selection */}
        <div className="space-y-2">
          <p className="text-sm font-medium">Step 2 — Select CSV file</p>
          <input
            type="file"
            accept=".csv"
            onChange={handleFileChange}
            className="block w-full text-sm text-muted-foreground file:mr-4 file:py-1.5 file:px-3 file:rounded-md file:border file:border-input file:bg-background file:text-sm file:font-medium file:cursor-pointer hover:file:bg-accent"
          />
          {current.file && (
            <p className="text-xs text-muted-foreground">
              Selected: <span className="font-medium text-foreground">{current.file.name}</span>{' '}
              ({(current.file.size / 1024).toFixed(1)} KB)
            </p>
          )}
        </div>

        {/* CSV preview */}
        {current.csvText && (
          <div className="space-y-1.5">
            <p className="text-sm font-medium">Preview</p>
            <ImportPreviewTable csvText={current.csvText} />
          </div>
        )}

        <hr className="border-border" />

        {/* Import action */}
        <div className="flex items-center gap-3">
          <Button
            onClick={handleImport}
            disabled={!current.file || mutation.isPending}
          >
            {mutation.isPending ? (
              <span className="flex items-center gap-2">
                <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Importing…
              </span>
            ) : (
              'Import'
            )}
          </Button>
          {errorMessage && (
            <p className="text-sm text-destructive">{errorMessage}</p>
          )}
        </div>

        {/* Results */}
        {current.result && (
          <>
            <hr className="border-border" />
            <div className="space-y-1.5">
              <p className="text-sm font-medium">Results</p>
              <ImportResults result={current.result} typeLabel={activeTab} />
            </div>
          </>
        )}
      </Card>
    </div>
  )
}
