import * as React from 'react'
import type { ToastProps } from '@/components/ui/toast'

const TOAST_LIMIT = 5
const TOAST_REMOVE_DELAY = 4000

type ToastItem = ToastProps & {
  id: string
  title?: React.ReactNode
  description?: React.ReactNode
  action?: React.ReactElement
}

type State = { toasts: ToastItem[] }

let count = 0
function genId() {
  count = (count + 1) % Number.MAX_SAFE_INTEGER
  return count.toString()
}

type Action =
  | { type: 'ADD_TOAST'; toast: ToastItem }
  | { type: 'UPDATE_TOAST'; toast: Partial<ToastItem> & { id: string } }
  | { type: 'DISMISS_TOAST'; toastId?: string }
  | { type: 'REMOVE_TOAST'; toastId?: string }

const toastTimeouts = new Map<string, ReturnType<typeof setTimeout>>()
const listeners: Array<(state: State) => void> = []
let memoryState: State = { toasts: [] }

function addToRemoveQueue(toastId: string) {
  if (toastTimeouts.has(toastId)) return
  const timeout = setTimeout(() => {
    toastTimeouts.delete(toastId)
    dispatch({ type: 'REMOVE_TOAST', toastId })
  }, TOAST_REMOVE_DELAY)
  toastTimeouts.set(toastId, timeout)
}

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'ADD_TOAST':
      return { ...state, toasts: [action.toast, ...state.toasts].slice(0, TOAST_LIMIT) }
    case 'UPDATE_TOAST':
      return {
        ...state,
        toasts: state.toasts.map((t) => (t.id === action.toast.id ? { ...t, ...action.toast } : t)),
      }
    case 'DISMISS_TOAST': {
      if (action.toastId) addToRemoveQueue(action.toastId)
      else state.toasts.forEach((t) => addToRemoveQueue(t.id))
      return {
        ...state,
        toasts: state.toasts.map((t) =>
          !action.toastId || t.id === action.toastId ? { ...t, open: false } : t,
        ),
      }
    }
    case 'REMOVE_TOAST':
      return {
        ...state,
        toasts: action.toastId ? state.toasts.filter((t) => t.id !== action.toastId) : [],
      }
  }
}

function dispatch(action: Action) {
  memoryState = reducer(memoryState, action)
  listeners.forEach((l) => l(memoryState))
}

type Toast = Omit<ToastItem, 'id'>

function toast(props: Toast) {
  const id = genId()
  const dismiss = () => dispatch({ type: 'DISMISS_TOAST', toastId: id })
  dispatch({
    type: 'ADD_TOAST',
    toast: { ...props, id, open: true, onOpenChange: (open) => { if (!open) dismiss() } },
  })
  return { id, dismiss, update: (p: ToastItem) => dispatch({ type: 'UPDATE_TOAST', toast: { ...p, id } }) }
}

function useToast() {
  const [state, setState] = React.useState<State>(memoryState)
  React.useEffect(() => {
    listeners.push(setState)
    return () => {
      const idx = listeners.indexOf(setState)
      if (idx > -1) listeners.splice(idx, 1)
    }
  }, [])
  return { ...state, toast, dismiss: (id?: string) => dispatch({ type: 'DISMISS_TOAST', toastId: id }) }
}

export { useToast, toast }
