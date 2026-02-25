import { useEffect, useRef, useState } from 'react'
import { ENDPOINTS } from '@/lib/api/endpoints'
import type { Schedule } from '../types'

// SSE event types emitted by the schedule generation stream
type StreamStatus = 'idle' | 'started' | 'completed' | 'failed'

interface ProgressPayload {
  assigned: number
  total: number
}

interface UseScheduleStreamReturn {
  status: StreamStatus
  progress: ProgressPayload | null
  schedule: Schedule | null
}

const BASE_URL = import.meta.env.VITE_API_URL || '/api'

/**
 * useScheduleStream opens an SSE connection to track real-time schedule
 * generation progress. Only connects when scheduleId is non-null.
 * Auto-closes on terminal events (completed/failed) and on unmount.
 */
export function useScheduleStream(scheduleId: string | null): UseScheduleStreamReturn {
  const [status, setStatus] = useState<StreamStatus>('idle')
  const [progress, setProgress] = useState<ProgressPayload | null>(null)
  const [schedule, setSchedule] = useState<Schedule | null>(null)

  // Track EventSource ref for cleanup
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    // No ID yet — stay idle
    if (!scheduleId) {
      setStatus('idle')
      setProgress(null)
      setSchedule(null)
      return
    }

    const token = localStorage.getItem('access_token') ?? ''
    const url = `${BASE_URL}${ENDPOINTS.timetable.scheduleStream(scheduleId)}?token=${encodeURIComponent(token)}`

    const es = new EventSource(url)
    esRef.current = es

    // Generic helper to close and stop reconnecting
    const closeStream = () => {
      es.close()
      esRef.current = null
    }

    es.addEventListener('started', () => {
      setStatus('started')
      setProgress(null)
    })

    es.addEventListener('progress', (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data as string) as ProgressPayload
        setProgress(payload)
      } catch {
        // Ignore malformed progress frames
      }
    })

    es.addEventListener('completed', (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data as string) as Schedule
        setSchedule(payload)
      } catch {
        // Completed without parseable body still marks done
      }
      setStatus('completed')
      closeStream()
    })

    es.addEventListener('failed', () => {
      setStatus('failed')
      closeStream()
    })

    es.onerror = () => {
      // Browser will attempt reconnect automatically; if already terminal, close
      if (status === 'completed' || status === 'failed') {
        closeStream()
      }
    }

    return () => {
      closeStream()
    }
    // Re-run only when scheduleId changes; status intentionally excluded to avoid
    // re-opening the stream when terminal state updates it
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scheduleId])

  return { status, progress, schedule }
}
