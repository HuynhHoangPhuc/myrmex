# Myrmex ERP — User Guide

## Overview

Myrmex is a university faculty management system covering HR, Subjects, Timetable, and Student records. Access is role-based; what you see depends on your role.

## Roles at a Glance

| Role | Key Capabilities |
|------|-----------------|
| Super Admin | Full system access, all admin functions |
| Admin | User management, bulk import, audit logs, role assignment |
| Department Head | Department teachers, subjects, semester setup, schedule generation |
| Teacher | View schedule, set availability, enter grades |
| Student | View schedule, enrolled subjects, transcript, enrollment requests |

## Login Methods

| Method | Available To |
|--------|-------------|
| Email + Password | All roles |
| Google OAuth (SSO) | Teachers, Admins |
| Microsoft OAuth (SSO) | Students (invite-code registration) |

## Quick Start by Role

1. **Admin** — Log in → change default password → bulk import users → assign roles. See [Admin Guide](admin-guide.md).
2. **Department Head** — Log in → set up departments → create subjects → configure semester → generate schedule. See [Dept Head Guide](department-head-guide.md).
3. **Teacher** — Log in → view assigned schedule → set weekly availability → enter grades. See [Teacher Guide](teacher-guide.md).
4. **Student** — Register via invite code or Microsoft SSO → view schedule → request enrollments. See [Student Guide](student-guide.md).

## Navigation Map

```
/dashboard          — Home overview (all roles)
/hr                 — Teachers & Departments (Admin, Dept Head)
/subjects           — Subject catalog, prerequisites, offerings
/timetable          — Semesters, schedules, generation (Admin, Dept Head)
/students           — Student list, enrollments (Admin)
/grades             — Grade entry (Teacher)
/analytics          — Workload & utilization charts
/notifications      — In-app notification center
/admin              — Role management, audit logs, bulk import (Admin only)
/help               — This guide
```

## Common Actions

- **Search**: Use the top search bar (Ctrl+K) for quick navigation.
- **Theme**: Toggle dark/light mode via the icon in the top-right header.
- **Notifications**: Bell icon shows unread count; click to view all.
- **Logout**: User menu in top-right corner → Sign out.

## Support

Contact your system administrator for account issues, password resets, or access problems.
