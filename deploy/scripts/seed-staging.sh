#!/bin/bash
# Seed staging database with realistic HCMUS-like data.
# Usage: ./deploy/scripts/seed-staging.sh <STAGING_DATABASE_URL>
# Example: ./deploy/scripts/seed-staging.sh "postgresql://user:pass@host/myrmex"
set -euo pipefail

DB_URL="${1:?Usage: seed-staging.sh <STAGING_DATABASE_URL>}"

echo "=== Seeding staging database ==="
psql "$DB_URL" <<'SQL'

-- Departments (HCMUS-like structure)
INSERT INTO hr.departments (id, name, code, description) VALUES
  (gen_random_uuid(), 'Computer Science',        'CS',   'Faculty of Computer Science'),
  (gen_random_uuid(), 'Information Technology',  'IT',   'Faculty of Information Technology'),
  (gen_random_uuid(), 'Mathematics & Statistics', 'MTH', 'Faculty of Mathematics and Statistics')
ON CONFLICT DO NOTHING;

-- Super admin (password: Staging@2026 — bcrypt hash)
INSERT INTO core.users (id, email, password_hash, full_name, role, department_id)
SELECT
  gen_random_uuid(),
  'admin@staging.myrmex.local',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LEnBjhV0h5u',
  'Staging Admin',
  'super_admin',
  NULL
WHERE NOT EXISTS (SELECT 1 FROM core.users WHERE email = 'admin@staging.myrmex.local');

-- Sample teachers (20) — realistic Vietnamese names
WITH dept AS (SELECT id, code FROM hr.departments)
INSERT INTO hr.teachers (id, user_id, employee_code, full_name, email, department_id, max_hours_per_week)
SELECT
  gen_random_uuid(),
  gen_random_uuid(),
  'GV' || lpad(n::text, 4, '0'),
  name,
  lower(replace(name, ' ', '.')) || '@hcmus.edu.vn',
  (SELECT id FROM dept ORDER BY random() LIMIT 1),
  12
FROM (VALUES
  (1, 'Nguyen Van An'),   (2, 'Tran Thi Bich'),  (3, 'Le Van Cuong'),
  (4, 'Pham Thi Dung'),   (5, 'Hoang Van Em'),   (6, 'Vo Thi Phuong'),
  (7, 'Dang Van Giang'),  (8, 'Bui Thi Hang'),   (9, 'Do Van Ich'),
  (10, 'Ngo Thi Kim'),    (11, 'Dinh Van Long'),  (12, 'Ly Thi Mai'),
  (13, 'Truong Van Nam'), (14, 'Duong Thi Oanh'), (15, 'Mai Van Phuc'),
  (16, 'Luong Thi Que'),  (17, 'Cao Van Rong'),   (18, 'Tang Thi Son'),
  (19, 'Lam Van Thanh'),  (20, 'Trinh Thi Uyen')
) AS t(n, name)
ON CONFLICT DO NOTHING;

-- Sample subjects (15)
WITH dept AS (SELECT id, code FROM hr.departments WHERE code = 'CS' LIMIT 1)
INSERT INTO subject.subjects (id, code, name, credits, weekly_hours, department_id, description)
VALUES
  (gen_random_uuid(), 'CS101', 'Introduction to Programming',  3, 4, (SELECT id FROM dept), 'Python basics'),
  (gen_random_uuid(), 'CS201', 'Data Structures',              3, 4, (SELECT id FROM dept), 'Arrays, lists, trees'),
  (gen_random_uuid(), 'CS301', 'Algorithms',                   3, 4, (SELECT id FROM dept), 'Sorting, searching, graph algorithms'),
  (gen_random_uuid(), 'CS401', 'Database Systems',             3, 4, (SELECT id FROM dept), 'SQL, normalization, transactions'),
  (gen_random_uuid(), 'CS501', 'Software Engineering',         3, 4, (SELECT id FROM dept), 'SDLC, patterns, testing'),
  (gen_random_uuid(), 'CS601', 'Computer Networks',            3, 4, (SELECT id FROM dept), 'TCP/IP, HTTP, security'),
  (gen_random_uuid(), 'CS701', 'Operating Systems',            3, 4, (SELECT id FROM dept), 'Processes, memory, I/O'),
  (gen_random_uuid(), 'CS801', 'Machine Learning',             3, 4, (SELECT id FROM dept), 'Supervised + unsupervised learning'),
  (gen_random_uuid(), 'CS901', 'Web Development',              3, 4, (SELECT id FROM dept), 'HTML/CSS/JS, React, backend APIs'),
  (gen_random_uuid(), 'CS201L', 'Data Structures Lab',        1, 2, (SELECT id FROM dept), 'Practical lab for CS201')
ON CONFLICT DO NOTHING;

-- Prerequisite: CS201 requires CS101
INSERT INTO subject.prerequisites (subject_id, prerequisite_id, type)
SELECT s.id, p.id, 'required'
FROM subject.subjects s, subject.subjects p
WHERE s.code = 'CS201' AND p.code = 'CS101'
ON CONFLICT DO NOTHING;

-- Prerequisite: CS301 requires CS201
INSERT INTO subject.prerequisites (subject_id, prerequisite_id, type)
SELECT s.id, p.id, 'required'
FROM subject.subjects s, subject.subjects p
WHERE s.code = 'CS301' AND p.code = 'CS201'
ON CONFLICT DO NOTHING;

-- Sample students (50) — realistic student codes
WITH dept AS (SELECT id FROM hr.departments WHERE code = 'CS' LIMIT 1)
INSERT INTO student.students (id, user_id, student_code, full_name, email, department_id)
SELECT
  gen_random_uuid(),
  gen_random_uuid(),
  '22' || lpad(n::text, 6, '0'),
  'Student ' || n,
  'student' || n || '@student.hcmus.edu.vn',
  (SELECT id FROM dept)
FROM generate_series(1, 50) AS n
ON CONFLICT DO NOTHING;

-- Semester
INSERT INTO timetable.semesters (id, name, start_date, end_date, is_active)
VALUES (gen_random_uuid(), 'Semester 2 2025-2026', '2026-01-15', '2026-05-31', true)
ON CONFLICT DO NOTHING;

-- Rooms
INSERT INTO timetable.rooms (id, name, capacity, building)
VALUES
  (gen_random_uuid(), 'B1.01', 50, 'B1'),
  (gen_random_uuid(), 'B1.02', 50, 'B1'),
  (gen_random_uuid(), 'B2.01', 80, 'B2'),
  (gen_random_uuid(), 'B2.02', 80, 'B2'),
  (gen_random_uuid(), 'Lab.01', 30, 'Lab')
ON CONFLICT DO NOTHING;

SQL

echo "=== Staging seed complete ==="
