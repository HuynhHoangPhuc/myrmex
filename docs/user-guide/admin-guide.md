# Admin Guide

## First Login & Password Change

1. Receive credentials (email + temporary password) from Super Admin.
2. Navigate to the login page → enter email and temporary password → click **Sign In**.
3. You will be prompted to change your password immediately.
4. Enter new password (min 8 characters, at least one uppercase, one digit) → **Save**.
5. You are now on the Dashboard.

> Google OAuth admins: click **Continue with Google** instead of email/password.

---

## Bulk User Import

Admins can import teachers and students in batches via CSV upload.

### Access
Navigate to **Admin → Bulk Import** (`/admin/import`).

### Tabs
- **Teachers tab** — imports users with role `teacher`
- **Students tab** — imports users with role `student`

### Steps

1. Click **Download Template** to get the correct CSV format for the selected tab.
2. Fill in the CSV:
   - Teachers: `email, full_name, department_id, employee_id`
   - Students: `email, full_name, student_id, program, year`
3. Click **Choose File** → select your CSV → click **Upload**.
4. The system validates each row. A progress bar shows import status.
5. When complete, a summary shows:
   - Total rows processed
   - Successfully imported count
   - Failed rows with error reasons
6. Click **Download Error Report** to get a CSV of failed rows with error details.
7. Fix errors in the report and re-upload only the failed rows.

### CSV Rules
- First row must be the header row (exact column names from template).
- Maximum 500 rows per upload.
- Duplicate emails are skipped with an error.
- `department_id` must match an existing department ID.

---

## User Management

Navigate to **HR → Teachers** or **Students** for user lists.

### List & Search
- Use the search bar to filter by name or email.
- Filter by department using the dropdown.
- Click a column header to sort.

### Edit a User
1. Click the user row or the **Edit** icon.
2. Modify fields: name, department, employee/student ID.
3. Click **Save**.

### Assign Roles
1. Navigate to **Admin → Role Management** (`/admin/roles`).
2. Search for the user by email or name.
3. Select the desired role from the dropdown.
4. Click **Assign Role** → confirm in the dialog.

> Role changes take effect on next login.

---

## Department Management

Navigate to **HR → Departments** (`/hr/departments`).

| Action | Steps |
|--------|-------|
| Create department | Click **New Department** → enter name, code → **Save** |
| Edit department | Click department row → edit fields → **Save** |
| Assign head | Edit department → select **Department Head** user → **Save** |
| Delete department | Click **Delete** → confirm (only if no teachers assigned) |

---

## Role Management

Navigate to **Admin → Role Management** (`/admin/roles`).

- Roles available: `student`, `teacher`, `dept_head`, `admin`, `super_admin`
- Admins can assign up to `admin` role; only Super Admin can assign `super_admin`.
- A user can hold one role at a time.
- Filter the list by role type using the dropdown at the top.

---

## Audit Logs

Navigate to **Admin → Audit Logs** (`/admin/audit-logs`).

All system mutations are recorded immutably. Logs cannot be deleted.

### Filters

| Filter | Description |
|--------|-------------|
| Resource type | teacher, student, department, subject, role, etc. |
| Action | HTTP method or named action (e.g., `POST /teachers`) |
| User ID | UUID of the actor |

### Reading a Log Entry

Click any row to expand it and see:
- **Before** — JSON snapshot of the resource before the change
- **After** — JSON snapshot after the change
- **Resource ID** — UUID of the affected record

### Export
Currently logs are view-only in the UI. For bulk export, contact your DBA to query the `audit_logs` table directly.

---

## Broadcast Announcements

Navigate to **Notifications** (`/notifications`) → click **Broadcast**.

1. Enter announcement title and body text.
2. Select target audience: **All Users**, **Teachers only**, or **Students only**.
3. Click **Send** → confirm.

All targeted users receive an in-app notification immediately.
