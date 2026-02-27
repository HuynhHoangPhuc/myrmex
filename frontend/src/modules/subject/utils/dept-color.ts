// Deterministic color assignment for department IDs using a simple hash.
// Ensures consistent colors across renders without storing state.

const COLORS = [
  '#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6',
  '#ec4899', '#06b6d4', '#84cc16', '#f97316', '#6366f1',
]

export function getDeptColor(departmentId: string): string {
  let hash = 0
  for (let i = 0; i < departmentId.length; i++) {
    hash = ((hash << 5) - hash + departmentId.charCodeAt(i)) | 0
  }
  return COLORS[Math.abs(hash) % COLORS.length]
}
