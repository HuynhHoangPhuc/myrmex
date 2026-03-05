#!/bin/bash
# Post-import verification: counts, orphan checks, relationship integrity.
# Usage: ./deploy/migration/verify-import.sh <DATABASE_URL>
set -euo pipefail

DB_URL="${1:?Usage: verify-import.sh <DATABASE_URL>}"

echo "=== Import Verification ==="

echo ""
echo "--- Record counts ---"
psql "$DB_URL" -t <<'SQL'
SELECT 'departments'   AS entity, count(*) FROM hr.departments
UNION ALL SELECT 'teachers',  count(*) FROM hr.teachers
UNION ALL SELECT 'subjects',  count(*) FROM subject.subjects
UNION ALL SELECT 'prereqs',   count(*) FROM subject.prerequisites
UNION ALL SELECT 'students',  count(*) FROM student.students
UNION ALL SELECT 'users',     count(*) FROM core.users
UNION ALL SELECT 'rooms',     count(*) FROM timetable.rooms
UNION ALL SELECT 'semesters', count(*) FROM timetable.semesters
ORDER BY 1;
SQL

echo ""
echo "--- Orphaned teachers (no valid department) ---"
ORPHANED_TEACHERS=$(psql "$DB_URL" -t -c "
  SELECT count(*) FROM hr.teachers t
  LEFT JOIN hr.departments d ON t.department_id = d.id
  WHERE d.id IS NULL;")
echo "  Orphaned teachers: $ORPHANED_TEACHERS"

echo ""
echo "--- Orphaned students (no valid department) ---"
ORPHANED_STUDENTS=$(psql "$DB_URL" -t -c "
  SELECT count(*) FROM student.students s
  LEFT JOIN hr.departments d ON s.department_id = d.id
  WHERE d.id IS NULL;")
echo "  Orphaned students: $ORPHANED_STUDENTS"

echo ""
echo "--- Subjects with broken prerequisites ---"
psql "$DB_URL" -t <<'SQL'
SELECT s.code, p.code AS missing_prereq
FROM subject.prerequisites sp
JOIN subject.subjects s  ON sp.subject_id = s.id
JOIN subject.subjects p  ON sp.prerequisite_id = p.id
WHERE p.id IS NULL
LIMIT 10;
SQL

echo ""
echo "--- Users without role ---"
NOROLE=$(psql "$DB_URL" -t -c "SELECT count(*) FROM core.users WHERE role IS NULL OR role = '';")
echo "  Users without role: $NOROLE"

echo ""
echo "--- Super admins ---"
psql "$DB_URL" -t -c "SELECT email, full_name FROM core.users WHERE role = 'super_admin';"

FAIL=0
[ "$ORPHANED_TEACHERS" -gt 0 ] && echo "FAIL: orphaned teachers" && FAIL=1
[ "$ORPHANED_STUDENTS" -gt 0 ] && echo "FAIL: orphaned students" && FAIL=1
[ "$NOROLE" -gt 0 ] && echo "FAIL: users without role" && FAIL=1

echo ""
if [ "$FAIL" -eq 0 ]; then
  echo "✅ Verification passed — data integrity OK."
else
  echo "❌ Verification FAILED — fix issues before go-live."
  exit 1
fi
