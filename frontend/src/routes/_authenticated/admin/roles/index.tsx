import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { PageHeader } from '@/components/shared/page-header'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toast } from '@/lib/hooks/use-toast'
import type { User, UserRole } from '@/lib/api/types'
import { useAllDepartments } from '@/modules/hr/hooks/use-departments'

export const Route = createFileRoute('/_authenticated/admin/roles/')({
  component: RoleManagementPage,
})

const ALL_ROLES: UserRole[] = [
  'super_admin',
  'admin',
  'dean',
  'dept_head',
  'manager',
  'viewer',
  'student',
  'teacher',
]

// Roles that require a department assignment
const DEPT_SCOPED_ROLES: UserRole[] = ['dept_head', 'teacher']

const ROLE_BADGE_VARIANT: Record<UserRole, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  super_admin: 'destructive',
  admin: 'destructive',
  dean: 'default',
  dept_head: 'default',
  manager: 'secondary',
  viewer: 'outline',
  student: 'secondary',
  teacher: 'secondary',
}

function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const { data } = await apiClient.get<{ users: User[]; total: number }>(
        `${ENDPOINTS.users.list}?page_size=100`,
      )
      return data.users
    },
  })
}

function useUpdateUserRole() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      userId,
      role,
      department_id,
    }: {
      userId: string
      role: string
      department_id?: string
    }) => {
      const { data } = await apiClient.patch<User>(ENDPOINTS.users.updateRole(userId), {
        role,
        department_id,
      })
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast({ title: 'Role updated successfully' })
    },
    onError: (err: Error) => {
      toast({ title: 'Failed to update role', description: err.message, variant: 'destructive' })
    },
  })
}

interface EditState {
  user: User
  role: UserRole
  departmentId: string
}

function RoleManagementPage() {
  const { data: users = [], isLoading } = useUsers()
  const { data: departments = [] } = useAllDepartments()
  const updateRole = useUpdateUserRole()
  const [search, setSearch] = React.useState('')
  const [editing, setEditing] = React.useState<EditState | null>(null)

  const filtered = users.filter(
    (u) =>
      u.email.toLowerCase().includes(search.toLowerCase()) ||
      u.full_name.toLowerCase().includes(search.toLowerCase()),
  )

  function openEdit(user: User) {
    setEditing({ user, role: user.role, departmentId: user.department_id ?? '' })
  }

  async function handleSave() {
    if (!editing) return
    const payload: { userId: string; role: string; department_id?: string } = {
      userId: editing.user.id,
      role: editing.role,
    }
    if (editing.departmentId) {
      payload.department_id = editing.departmentId
    }
    await updateRole.mutateAsync(payload)
    setEditing(null)
  }

  const needsDept = editing ? DEPT_SCOPED_ROLES.includes(editing.role) : false

  return (
    <div className="space-y-6">
      <PageHeader
        title="Role Management"
        description="Assign roles and department scopes to users"
      />

      <div className="flex items-center gap-3">
        <Input
          placeholder="Search by name or email..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-sm"
        />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Department</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-[80px]" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                  Loading users...
                </TableCell>
              </TableRow>
            ) : filtered.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                  No users found
                </TableCell>
              </TableRow>
            ) : (
              filtered.map((user) => {
                const dept = departments.find((d) => d.id === user.department_id)
                return (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.full_name}</TableCell>
                    <TableCell className="text-muted-foreground">{user.email}</TableCell>
                    <TableCell>
                      <Badge variant={ROLE_BADGE_VARIANT[user.role] ?? 'outline'}>
                        {user.role}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {dept?.name ?? (user.department_id ? user.department_id : '—')}
                    </TableCell>
                    <TableCell>
                      <Badge variant={user.is_active !== false ? 'default' : 'outline'}>
                        {user.is_active !== false ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button size="sm" variant="outline" onClick={() => openEdit(user)}>
                        Edit Role
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      {/* Edit Role Dialog */}
      <Dialog open={!!editing} onOpenChange={(open) => !open && setEditing(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Edit Role — {editing?.user.full_name}</DialogTitle>
          </DialogHeader>

          {editing && (
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label>Role</Label>
                <select
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  value={editing.role}
                  onChange={(e) => setEditing({ ...editing, role: e.target.value as UserRole })}
                >
                  {ALL_ROLES.map((r) => (
                    <option key={r} value={r}>
                      {r}
                    </option>
                  ))}
                </select>
              </div>

              {needsDept && (
                <div className="space-y-2">
                  <Label>Department</Label>
                  <select
                    className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={editing.departmentId}
                    onChange={(e) => setEditing({ ...editing, departmentId: e.target.value })}
                  >
                    <option value="">— Select department —</option>
                    {departments.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name}
                      </option>
                    ))}
                  </select>
                  {needsDept && !editing.departmentId && (
                    <p className="text-xs text-destructive">
                      Department is required for {editing.role} role
                    </p>
                  )}
                </div>
              )}

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={() => setEditing(null)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleSave}
                  disabled={updateRole.isPending || (needsDept && !editing.departmentId)}
                >
                  {updateRole.isPending ? 'Saving...' : 'Save Role'}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
