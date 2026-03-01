import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Transcript } from '@/modules/student/types'

export function useMyTranscript() {
  return useQuery({
    queryKey: ['student-portal', 'transcript'] as const,
    queryFn: async () => {
      const { data } = await apiClient.get<Transcript>(ENDPOINTS.studentPortal.myTranscript)
      return data
    },
  })
}

// Triggers a browser download of the PDF transcript
export function downloadTranscript() {
  const token = localStorage.getItem('access_token')
  const url = `${import.meta.env.VITE_API_URL ?? '/api'}${ENDPOINTS.studentPortal.exportTranscript}`
  const a = document.createElement('a')
  a.href = url
  a.download = 'transcript.pdf'
  // Append token as query param since fetch-based download needs auth
  a.href = `${url}?token=${token ?? ''}`
  a.click()
}
