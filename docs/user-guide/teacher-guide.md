# Teacher Guide

## Login

### Email + Password
1. Go to the login page.
2. Enter your institutional email and password → click **Sign In**.
3. If it is your first login, you will be prompted to set a new password.

### Google OAuth (SSO)
1. Click **Continue with Google**.
2. Select your institutional Google account.
3. You are redirected to the Dashboard on success.

> If you see "Account not found", contact your admin — your email must be pre-registered before OAuth login works.

---

## Viewing Your Assigned Schedule

1. Navigate to **Timetable → Schedules** (`/timetable/schedules`).
2. Your schedule is shown filtered to your assigned sessions by default.
3. Use the **Semester** dropdown to switch between semesters.
4. Each session shows: subject name, room, day, start/end time, student count.
5. Click a session row to see full details including enrolled students.

---

## Setting Weekly Availability Preferences

Availability preferences inform the schedule generator. Preferences are per-semester.

1. Navigate to **Timetable → Schedules** → click **My Availability**.
2. Select the target semester from the dropdown.
3. A weekly grid is shown (Mon–Sat, time slots).
4. Click cells to toggle **Available** (green) / **Unavailable** (grey).
   - Green = you prefer to teach at this time.
   - Grey = you prefer not to teach at this time (system may still assign if necessary).
5. Click **Save Preferences**.

> Preferences must be submitted before the schedule generation deadline set by the Department Head.

---

## Entering Student Grades

1. Navigate to **Grades** (`/grades`).
2. Select the **Semester** and **Subject** from the dropdowns.
3. The enrolled student list for that subject appears.
4. Enter grade values in the **Grade** column (numeric, 0–10 scale or as configured).
5. Optionally add a **Note** per student.
6. Click **Save Grades** at the bottom.
   - Saved grades are visible to students immediately.
   - You can edit grades until the semester grading window closes.

### Grade Status Indicators

| Status | Meaning |
|--------|---------|
| Draft | Entered but not yet published |
| Published | Visible to student |
| Locked | Grading window closed; contact admin to unlock |

---

## Notification Preferences

1. Navigate to **Notifications** (`/notifications`) → click **Preferences** tab.
2. Toggle on/off:
   - Schedule changes
   - New enrollment requests
   - Grading deadline reminders
   - System announcements
3. Click **Save Preferences**.

---

## AI Chat Assistant

The AI assistant can answer questions about system features, policies, and your data.

1. Click the **Chat** icon in the bottom-right corner of any page.
2. Type your question in natural language.
   - Example: "Show me my schedule for next week"
   - Example: "How many students are enrolled in CS101?"
3. The assistant responds with data pulled from your account context.
4. Chat history persists within your session; it resets on logout.

> The AI assistant cannot modify data — it is read-only and advisory only.
