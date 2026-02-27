-- Myrmex demo seed data
-- Comprehensive seed demonstrating all modules, features, and complex dependencies
-- Idempotent: uses ON CONFLICT DO NOTHING / DO UPDATE
-- Run: make seed

BEGIN;

-- =====================================================================
-- 0. Core: Demo admin user
--    email: admin@myrmex.dev  password: demo1234
-- =====================================================================
INSERT INTO core.users (id, email, password_hash, full_name, role, is_active)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'admin@myrmex.dev',
  '$2b$12$SXaQW37wRUZMFheHe/xhPeg0yN/EC0D89uXN1h9G4Ltvv/8rqKyNe',
  'Demo Admin',
  'admin',
  true
)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 1. HR: Departments (3)
-- =====================================================================
INSERT INTO hr.departments (id, name, code)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'Computer Science', 'CS'),
  ('00000000-0000-0000-0000-000000000002', 'Mathematics',      'MATH'),
  ('00000000-0000-0000-0000-000000000003', 'Physics',          'PHYS')
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 2. HR: Teachers (11 — 10 active, 1 inactive for filter demo)
-- =====================================================================
INSERT INTO hr.teachers (id, employee_code, full_name, email, department_id, max_hours_per_week, title, is_active)
VALUES
  ('00000000-0000-0001-0000-000000000001', 'EMP001', 'Alice Nguyen',  'alice@myrmex.dev',  '00000000-0000-0000-0000-000000000001', 20, 'Dr.',  true),
  ('00000000-0000-0001-0000-000000000002', 'EMP002', 'Bob Tran',      'bob@myrmex.dev',    '00000000-0000-0000-0000-000000000001', 18, 'MSc.', true),
  ('00000000-0000-0001-0000-000000000003', 'EMP003', 'Carol Le',      'carol@myrmex.dev',  '00000000-0000-0000-0000-000000000002', 20, 'Dr.',  true),
  ('00000000-0000-0001-0000-000000000004', 'EMP004', 'David Pham',    'david@myrmex.dev',  '00000000-0000-0000-0000-000000000002', 16, 'MSc.', true),
  ('00000000-0000-0001-0000-000000000005', 'EMP005', 'Eve Hoang',     'eve@myrmex.dev',    '00000000-0000-0000-0000-000000000003', 20, 'Dr.',  true),
  ('00000000-0000-0001-0000-000000000006', 'EMP006', 'Frank Do',      'frank@myrmex.dev',  '00000000-0000-0000-0000-000000000003', 18, 'MSc.', true),
  ('00000000-0000-0001-0000-000000000007', 'EMP007', 'Grace Vu',      'grace@myrmex.dev',  '00000000-0000-0000-0000-000000000001', 16, 'MSc.', true),
  ('00000000-0000-0001-0000-000000000008', 'EMP008', 'Henry Ly',      'henry@myrmex.dev',  '00000000-0000-0000-0000-000000000001', 20, 'Dr.',  true),
  ('00000000-0000-0001-0000-000000000009', 'EMP009', 'Iris Dao',      'iris@myrmex.dev',   '00000000-0000-0000-0000-000000000001', 18, 'MSc.', true),
  ('00000000-0000-0001-0000-000000000010', 'EMP010', 'Jack Mai',      'jack@myrmex.dev',   '00000000-0000-0000-0000-000000000001', 20, 'Dr.',  true),
  -- Inactive teacher for filter demo
  ('00000000-0000-0001-0000-000000000011', 'EMP011', 'Kim Bui',       'kim@myrmex.dev',    '00000000-0000-0000-0000-000000000001', 20, 'MSc.', false)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 3. HR: Teacher specializations
-- =====================================================================
INSERT INTO hr.teacher_specializations (teacher_id, specialization)
VALUES
  -- Alice Nguyen: Programming, OOP, Software Engineering
  ('00000000-0000-0001-0000-000000000001', 'PROG'),
  ('00000000-0000-0001-0000-000000000001', 'OOP'),
  ('00000000-0000-0001-0000-000000000001', 'SE'),
  -- Bob Tran: Data Structures, Databases
  ('00000000-0000-0001-0000-000000000002', 'DS'),
  ('00000000-0000-0001-0000-000000000002', 'DB'),
  -- Carol Le: Calculus, Numerical Methods
  ('00000000-0000-0001-0000-000000000003', 'CALC'),
  ('00000000-0000-0001-0000-000000000003', 'NUM'),
  -- David Pham: Statistics, Linear Algebra
  ('00000000-0000-0001-0000-000000000004', 'STAT'),
  ('00000000-0000-0001-0000-000000000004', 'LA'),
  -- Eve Hoang: Mechanics, Quantum Mechanics
  ('00000000-0000-0001-0000-000000000005', 'MECH'),
  ('00000000-0000-0001-0000-000000000005', 'QM'),
  -- Frank Do: Electromagnetism
  ('00000000-0000-0001-0000-000000000006', 'EM'),
  ('00000000-0000-0001-0000-000000000006', 'PHYS'),
  -- Grace Vu: Networks, Security, Distributed Systems
  ('00000000-0000-0001-0000-000000000007', 'NET'),
  ('00000000-0000-0001-0000-000000000007', 'SEC'),
  ('00000000-0000-0001-0000-000000000007', 'DIST'),
  -- Henry Ly: Algorithms, Discrete Math
  ('00000000-0000-0001-0000-000000000008', 'ALGO'),
  ('00000000-0000-0001-0000-000000000008', 'DISC'),
  -- Iris Dao: Machine Learning, AI, Compilers
  ('00000000-0000-0001-0000-000000000009', 'ML'),
  ('00000000-0000-0001-0000-000000000009', 'AI'),
  ('00000000-0000-0001-0000-000000000009', 'COMP'),
  -- Jack Mai: Operating Systems, Capstone
  ('00000000-0000-0001-0000-000000000010', 'OS'),
  ('00000000-0000-0001-0000-000000000010', 'CAP')
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 4. HR: Teacher availability (varied — NOT all 100%)
--    Mon=1..Fri=5, periods 1-6. Each row: (teacher, day, start, end=start+1)
-- =====================================================================

-- Alice Nguyen: Mon-Thu full (1-6), Fri morning only (1-3)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000001'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(3),(4)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
UNION ALL
SELECT '00000000-0000-0001-0000-000000000001'::uuid, 5, p.period, p.period + 1
FROM (VALUES (1),(2),(3)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Bob Tran: Mon-Fri, skip period 1 on Mon & Wed (faculty meetings)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000002'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(3),(4),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
WHERE NOT (d.day IN (1,3) AND p.period = 1)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Carol Le: Mon-Thu full, Fri off
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000003'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(3),(4)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- David Pham: Tue-Fri only (Mon off for research)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000004'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (2),(3),(4),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Eve Hoang: Mon/Wed/Fri only (Tue/Thu research days)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000005'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(3),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Frank Do: Mon-Fri, periods 1-4 only (afternoon research)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000006'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(3),(4),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Grace Vu: Mon-Fri, no period 6 (picks up kids)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000007'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(3),(4),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Henry Ly: Mon-Thu full, Fri morning (1-3)
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000008'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(3),(4)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
UNION ALL
SELECT '00000000-0000-0001-0000-000000000008'::uuid, 5, p.period, p.period + 1
FROM (VALUES (1),(2),(3)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Iris Dao: Tue-Fri full, Mon off
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000009'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (2),(3),(4),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- Jack Mai: Mon/Tue/Thu/Fri, Wed off
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
SELECT '00000000-0000-0001-0000-000000000010'::uuid, d.day, p.period, p.period + 1
FROM (VALUES (1),(2),(4),(5)) AS d(day)
CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (teacher_id, day_of_week, start_period) DO NOTHING;

-- =====================================================================
-- 5. Subject: Subjects (25 — 24 active, 1 inactive)
--    Realistic CS/Math/Physics curriculum with varied credits & hours
-- =====================================================================
INSERT INTO subject.subjects (id, code, name, credits, description, department_id, weekly_hours, is_active)
VALUES
  -- Year 1: Foundation
  ('00000000-0000-0002-0000-000000000001', 'CS101',    'Intro to Programming',           3, 'Fundamentals of programming using Python. Variables, control flow, functions, basic OOP concepts.',                                '00000000-0000-0000-0000-000000000001', 4, true),
  ('00000000-0000-0002-0000-000000000002', 'MATH101',  'Calculus I',                     4, 'Limits, derivatives, integrals of single-variable functions. Applications to physics and engineering.',                              '00000000-0000-0000-0000-000000000002', 4, true),
  ('00000000-0000-0002-0000-000000000003', 'MATH102',  'Linear Algebra',                 3, 'Vectors, matrices, systems of equations, eigenvalues, vector spaces. Foundation for ML and numerical methods.',                     '00000000-0000-0000-0000-000000000002', 3, true),
  ('00000000-0000-0002-0000-000000000004', 'PHYS101',  'Mechanics',                      3, 'Newtonian mechanics: kinematics, dynamics, energy, momentum, rotational motion. Calculus-based.',                                   '00000000-0000-0000-0000-000000000003', 4, true),

  -- Year 2: Core
  ('00000000-0000-0002-0000-000000000005', 'CS201',    'Data Structures',                3, 'Arrays, linked lists, stacks, queues, trees, graphs, hash tables. Complexity analysis with Big-O notation.',                       '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000006', 'CS202',    'Object-Oriented Programming',    3, 'OOP principles: encapsulation, inheritance, polymorphism, design patterns. Java/C++ implementation.',                               '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000007', 'CS203',    'Discrete Mathematics',           3, 'Logic, sets, combinatorics, graph theory, recurrence relations. Mathematical foundation for algorithms.',                           '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000008', 'MATH201',  'Calculus II',                    4, 'Multivariable calculus, partial derivatives, multiple integrals, vector calculus. Stokes and Green theorems.',                       '00000000-0000-0000-0000-000000000002', 4, true),
  ('00000000-0000-0002-0000-000000000009', 'MATH202',  'Probability & Statistics',       3, 'Probability distributions, hypothesis testing, regression, Bayesian inference. R/Python for data analysis.',                        '00000000-0000-0000-0000-000000000002', 3, true),
  ('00000000-0000-0002-0000-000000000010', 'PHYS201',  'Electromagnetism',               3, 'Electric and magnetic fields, Maxwell equations, electromagnetic waves. Requires multivariable calculus.',                           '00000000-0000-0000-0000-000000000003', 4, true),

  -- Year 3: Intermediate
  ('00000000-0000-0002-0000-000000000011', 'CS301',    'Algorithms',                     3, 'Sorting, searching, dynamic programming, greedy algorithms, graph algorithms. NP-completeness introduction.',                      '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000012', 'CS302',    'Database Systems',               3, 'Relational model, SQL, normalization, indexing, transactions, query optimization. Hands-on with PostgreSQL.',                       '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000013', 'CS303',    'Computer Networks',              3, 'OSI/TCP-IP model, routing, transport protocols, network security fundamentals. Lab with Wireshark.',                                '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000014', 'CS304',    'Operating Systems',              3, 'Processes, threads, scheduling, memory management, file systems, concurrency. C implementation projects.',                          '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000015', 'CS305',    'Software Engineering',           3, 'SDLC, agile methodologies, requirements engineering, testing strategies, CI/CD. Team project.',                                     '00000000-0000-0000-0000-000000000001', 2, true),
  ('00000000-0000-0002-0000-000000000016', 'MATH301',  'Numerical Methods',              3, 'Root finding, interpolation, numerical integration, ODE solvers. Error analysis and stability. MATLAB/Python.',                     '00000000-0000-0000-0000-000000000002', 3, true),
  ('00000000-0000-0002-0000-000000000017', 'PHYS301',  'Quantum Mechanics',              3, 'Wave functions, Schrodinger equation, quantum states, measurement theory. Requires strong calculus background.',                    '00000000-0000-0000-0000-000000000003', 3, true),

  -- Year 4: Advanced / Specialization
  ('00000000-0000-0002-0000-000000000018', 'CS401',    'Machine Learning',               3, 'Supervised/unsupervised learning, neural networks, SVMs, decision trees. Math-heavy: requires stats + linear algebra.',            '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000019', 'CS402',    'Distributed Systems',            3, 'Consensus algorithms, replication, fault tolerance, CAP theorem. gRPC, message queues, distributed databases.',                     '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000020', 'CS403',    'Compiler Design',                3, 'Lexical analysis, parsing, semantic analysis, code generation, optimization. Build a compiler from scratch.',                       '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000021', 'CS404',    'Computer Security',              3, 'Cryptography, authentication, network security, web security, penetration testing. Ethical hacking labs.',                          '00000000-0000-0000-0000-000000000001', 2, true),
  ('00000000-0000-0002-0000-000000000022', 'CS405',    'Artificial Intelligence',        3, 'Search algorithms, knowledge representation, planning, NLP basics. Prolog and Python implementations.',                             '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000023', 'DB401',    'Advanced Databases',             3, 'NoSQL, distributed databases, data warehousing, stream processing. Hands-on with MongoDB, Redis, Kafka.',                           '00000000-0000-0000-0000-000000000001', 3, true),
  ('00000000-0000-0002-0000-000000000024', 'CS499',    'Capstone Project',               6, 'Industry-mentored team project applying full software engineering lifecycle. Deliverables: working product + documentation.',       '00000000-0000-0000-0000-000000000001', 6, true),

  -- Inactive subject for filter demo
  ('00000000-0000-0002-0000-000000000025', 'PHYS999',  'Physics Seminar',                2, 'Discontinued seminar series. Kept for historical records.',                                                                        '00000000-0000-0000-0000-000000000003', 2, false)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 6. Subject: Prerequisites (38 edges — 28 hard, 10 soft)
--    Demonstrates: diamonds, cross-dept deps, deep chains, fan-out, fan-in
-- =====================================================================
INSERT INTO subject.prerequisites (subject_id, prerequisite_id, type, priority)
VALUES
  -- ── Year 1 cross-dept ──
  -- PHYS101 soft-needs MATH101 (calculus-based mechanics)
  ('00000000-0000-0002-0000-000000000004', '00000000-0000-0002-0000-000000000002', 'soft', 3),

  -- ── Year 2: Core ──
  -- CS201 (Data Structures) ← CS101 (hard)
  ('00000000-0000-0002-0000-000000000005', '00000000-0000-0002-0000-000000000001', 'hard', 1),
  -- CS202 (OOP) ← CS101 (hard)
  ('00000000-0000-0002-0000-000000000006', '00000000-0000-0002-0000-000000000001', 'hard', 1),
  -- CS203 (Discrete Math) ← MATH101 (hard) — CROSS-DEPT: CS subject needs Math
  ('00000000-0000-0002-0000-000000000007', '00000000-0000-0002-0000-000000000002', 'hard', 1),
  -- MATH201 (Calc II) ← MATH101 (hard)
  ('00000000-0000-0002-0000-000000000008', '00000000-0000-0002-0000-000000000002', 'hard', 1),
  -- MATH202 (Prob & Stats) ← MATH101 (hard) + MATH102 (soft)
  ('00000000-0000-0002-0000-000000000009', '00000000-0000-0002-0000-000000000002', 'hard', 1),
  ('00000000-0000-0002-0000-000000000009', '00000000-0000-0002-0000-000000000003', 'soft', 3),
  -- PHYS201 (EM) ← PHYS101 (hard) + MATH201 (hard) — CROSS-DEPT convergence
  ('00000000-0000-0002-0000-000000000010', '00000000-0000-0002-0000-000000000004', 'hard', 1),
  ('00000000-0000-0002-0000-000000000010', '00000000-0000-0002-0000-000000000008', 'hard', 1),

  -- ── Year 3: Intermediate ──
  -- CS301 (Algorithms) ← CS201 (hard) + CS203 (hard) — DIAMOND pattern
  --   (CS101 → CS201) + (MATH101 → CS203) both feed CS301
  ('00000000-0000-0002-0000-000000000011', '00000000-0000-0002-0000-000000000005', 'hard', 1),
  ('00000000-0000-0002-0000-000000000011', '00000000-0000-0002-0000-000000000007', 'hard', 1),
  -- CS302 (Databases) ← CS201 (hard) + CS202 (soft)
  ('00000000-0000-0002-0000-000000000012', '00000000-0000-0002-0000-000000000005', 'hard', 1),
  ('00000000-0000-0002-0000-000000000012', '00000000-0000-0002-0000-000000000006', 'soft', 3),
  -- CS303 (Networks) ← CS201 (hard)
  ('00000000-0000-0002-0000-000000000013', '00000000-0000-0002-0000-000000000005', 'hard', 1),
  -- CS304 (OS) ← CS201 (hard) + CS202 (hard)
  ('00000000-0000-0002-0000-000000000014', '00000000-0000-0002-0000-000000000005', 'hard', 1),
  ('00000000-0000-0002-0000-000000000014', '00000000-0000-0002-0000-000000000006', 'hard', 1),
  -- CS305 (Software Eng) ← CS202 (hard) + CS302 (soft)
  ('00000000-0000-0002-0000-000000000015', '00000000-0000-0002-0000-000000000006', 'hard', 1),
  ('00000000-0000-0002-0000-000000000015', '00000000-0000-0002-0000-000000000012', 'soft', 2),
  -- MATH301 (Numerical Methods) ← MATH201 (hard) + MATH102 (hard) — DIAMOND
  --   MATH101 → MATH201 + MATH102 both feed MATH301
  ('00000000-0000-0002-0000-000000000016', '00000000-0000-0002-0000-000000000008', 'hard', 1),
  ('00000000-0000-0002-0000-000000000016', '00000000-0000-0002-0000-000000000003', 'hard', 1),
  -- PHYS301 (Quantum) ← PHYS201 (hard) + MATH201 (hard) — CROSS-DEPT convergence
  ('00000000-0000-0002-0000-000000000017', '00000000-0000-0002-0000-000000000010', 'hard', 1),
  ('00000000-0000-0002-0000-000000000017', '00000000-0000-0002-0000-000000000008', 'hard', 1),

  -- ── Year 4: Advanced ──
  -- CS401 (ML) ← CS301 (hard) + MATH202 (hard) + MATH301 (soft) — TRIPLE convergence
  ('00000000-0000-0002-0000-000000000018', '00000000-0000-0002-0000-000000000011', 'hard', 1),
  ('00000000-0000-0002-0000-000000000018', '00000000-0000-0002-0000-000000000009', 'hard', 1),
  ('00000000-0000-0002-0000-000000000018', '00000000-0000-0002-0000-000000000016', 'soft', 2),
  -- CS402 (Distributed) ← CS303 (hard) + CS304 (hard)
  ('00000000-0000-0002-0000-000000000019', '00000000-0000-0002-0000-000000000013', 'hard', 1),
  ('00000000-0000-0002-0000-000000000019', '00000000-0000-0002-0000-000000000014', 'hard', 1),
  -- CS403 (Compilers) ← CS301 (hard) + CS304 (hard) — needs both algo + OS knowledge
  ('00000000-0000-0002-0000-000000000020', '00000000-0000-0002-0000-000000000011', 'hard', 1),
  ('00000000-0000-0002-0000-000000000020', '00000000-0000-0002-0000-000000000014', 'hard', 1),
  -- CS404 (Security) ← CS303 (hard) + CS304 (soft)
  ('00000000-0000-0002-0000-000000000021', '00000000-0000-0002-0000-000000000013', 'hard', 1),
  ('00000000-0000-0002-0000-000000000021', '00000000-0000-0002-0000-000000000014', 'soft', 2),
  -- CS405 (AI) ← CS301 (hard) + MATH202 (hard) — CROSS-DEPT
  ('00000000-0000-0002-0000-000000000022', '00000000-0000-0002-0000-000000000011', 'hard', 1),
  ('00000000-0000-0002-0000-000000000022', '00000000-0000-0002-0000-000000000009', 'hard', 1),
  -- DB401 (Advanced DB) ← CS302 (hard) + CS402 (soft)
  ('00000000-0000-0002-0000-000000000023', '00000000-0000-0002-0000-000000000012', 'hard', 1),
  ('00000000-0000-0002-0000-000000000023', '00000000-0000-0002-0000-000000000019', 'soft', 2),
  -- CS499 (Capstone) ← CS305 (hard) + CS401 (soft)
  ('00000000-0000-0002-0000-000000000024', '00000000-0000-0002-0000-000000000015', 'hard', 1),
  ('00000000-0000-0002-0000-000000000024', '00000000-0000-0002-0000-000000000018', 'soft', 2)
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 7. Timetable: Rooms (7 — varied types and features)
-- =====================================================================
INSERT INTO timetable.rooms (id, name, capacity, type, features, is_active)
VALUES
  ('00000000-0000-0003-0000-000000000001', 'Room A101',    40, 'lecture', '{"projector","whiteboard"}',               true),
  ('00000000-0000-0003-0000-000000000002', 'Room A102',    40, 'lecture', '{"projector","whiteboard"}',               true),
  ('00000000-0000-0003-0000-000000000003', 'Room A103',    60, 'lecture', '{"projector","whiteboard","microphone"}',   true),
  ('00000000-0000-0003-0000-000000000004', 'Lab B201',     30, 'lab',     '{"computers","projector"}',                true),
  ('00000000-0000-0003-0000-000000000005', 'Lab B202',     25, 'lab',     '{"computers","projector"}',                true),
  ('00000000-0000-0003-0000-000000000006', 'Seminar C101', 20, 'seminar', '{"whiteboard"}',                           true),
  ('00000000-0000-0003-0000-000000000007', 'Hall D301',    80, 'lecture', '{"projector","microphone","recording"}',    true)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 8. Timetable: Semesters (2 — Fall 2025 + Spring 2026)
-- =====================================================================

-- Fall 2025: Foundation + Core subjects (12 subjects)
INSERT INTO timetable.semesters (id, name, year, term, start_date, end_date, offered_subject_ids)
VALUES (
  '00000000-0000-0004-0000-000000000001',
  'Fall 2025',
  2025,
  2,
  '2025-09-01',
  '2025-12-31',
  ARRAY[
    '00000000-0000-0002-0000-000000000001'::uuid,  -- CS101
    '00000000-0000-0002-0000-000000000002'::uuid,  -- MATH101
    '00000000-0000-0002-0000-000000000003'::uuid,  -- MATH102
    '00000000-0000-0002-0000-000000000004'::uuid,  -- PHYS101
    '00000000-0000-0002-0000-000000000005'::uuid,  -- CS201
    '00000000-0000-0002-0000-000000000006'::uuid,  -- CS202
    '00000000-0000-0002-0000-000000000007'::uuid,  -- CS203
    '00000000-0000-0002-0000-000000000008'::uuid,  -- MATH201
    '00000000-0000-0002-0000-000000000009'::uuid,  -- MATH202
    '00000000-0000-0002-0000-000000000010'::uuid,  -- PHYS201
    '00000000-0000-0002-0000-000000000011'::uuid,  -- CS301
    '00000000-0000-0002-0000-000000000013'::uuid   -- CS303
  ]
)
ON CONFLICT (id) DO NOTHING;

-- Spring 2026: Core + Advanced subjects (18 subjects)
INSERT INTO timetable.semesters (id, name, year, term, start_date, end_date, offered_subject_ids)
VALUES (
  '00000000-0000-0004-0000-000000000002',
  'Spring 2026',
  2026,
  1,
  '2026-01-06',
  '2026-05-31',
  ARRAY[
    '00000000-0000-0002-0000-000000000005'::uuid,  -- CS201
    '00000000-0000-0002-0000-000000000006'::uuid,  -- CS202
    '00000000-0000-0002-0000-000000000008'::uuid,  -- MATH201
    '00000000-0000-0002-0000-000000000011'::uuid,  -- CS301
    '00000000-0000-0002-0000-000000000012'::uuid,  -- CS302
    '00000000-0000-0002-0000-000000000013'::uuid,  -- CS303
    '00000000-0000-0002-0000-000000000014'::uuid,  -- CS304
    '00000000-0000-0002-0000-000000000015'::uuid,  -- CS305
    '00000000-0000-0002-0000-000000000016'::uuid,  -- MATH301
    '00000000-0000-0002-0000-000000000010'::uuid,  -- PHYS201
    '00000000-0000-0002-0000-000000000017'::uuid,  -- PHYS301
    '00000000-0000-0002-0000-000000000018'::uuid,  -- CS401
    '00000000-0000-0002-0000-000000000019'::uuid,  -- CS402
    '00000000-0000-0002-0000-000000000020'::uuid,  -- CS403
    '00000000-0000-0002-0000-000000000021'::uuid,  -- CS404
    '00000000-0000-0002-0000-000000000022'::uuid,  -- CS405
    '00000000-0000-0002-0000-000000000023'::uuid,  -- DB401
    '00000000-0000-0002-0000-000000000024'::uuid   -- CS499
  ]
)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 9. Timetable: Time slots — Mon(1)–Fri(5), periods 1–6 for BOTH semesters
-- =====================================================================

-- Fall 2025 time slots
INSERT INTO timetable.time_slots (id, semester_id, day_of_week, start_period, end_period)
SELECT
  ('00000000-0000-0005-' ||
   LPAD(CAST(d.day AS text), 4, '0') ||
   '-' ||
   LPAD(CAST(p.period AS text), 12, '0'))::uuid,
  '00000000-0000-0004-0000-000000000001',
  d.day,
  p.period,
  p.period + 1
FROM
  (VALUES (1),(2),(3),(4),(5))       AS d(day)
  CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (id) DO NOTHING;

-- Spring 2026 time slots (different UUID prefix to avoid conflicts)
INSERT INTO timetable.time_slots (id, semester_id, day_of_week, start_period, end_period)
SELECT
  ('00000000-0000-0006-' ||
   LPAD(CAST(d.day AS text), 4, '0') ||
   '-' ||
   LPAD(CAST(p.period AS text), 12, '0'))::uuid,
  '00000000-0000-0004-0000-000000000002',
  d.day,
  p.period,
  p.period + 1
FROM
  (VALUES (1),(2),(3),(4),(5))       AS d(day)
  CROSS JOIN (VALUES (1),(2),(3),(4),(5),(6)) AS p(period)
ON CONFLICT (id) DO NOTHING;

-- =====================================================================
-- 10. Timetable: Pre-generated schedule for Spring 2026 (draft)
--     18 entries — teacher assignments follow specializations
-- =====================================================================
INSERT INTO timetable.schedules (id, semester_id, name, status, score, hard_violations, soft_penalty, generated_at, created_at)
VALUES (
  '00000000-0000-0007-0000-000000000001',
  '00000000-0000-0004-0000-000000000002',
  'Spring 2026 — Auto-generated',
  'draft',
  85.5,
  0,
  14.5,
  '2026-01-05 10:00:00',
  '2026-01-05 10:00:00'
)
ON CONFLICT (id) DO NOTHING;

-- Schedule entries: subject → teacher → room → time slot
-- Time slot IDs for Spring 2026: 00000000-0000-0006-{day4}-{period12}
INSERT INTO timetable.schedule_entries (id, schedule_id, subject_id, teacher_id, room_id, time_slot_id, is_manual_override, subject_name, subject_code, teacher_name, department_id)
VALUES
  -- CS201 (Data Structures) → Bob Tran, Room A101, Mon period 2
  ('00000000-0000-0008-0000-000000000001', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000005', '00000000-0000-0001-0000-000000000002',
   '00000000-0000-0003-0000-000000000001', '00000000-0000-0006-0001-000000000002',
   false, 'Data Structures', 'CS201', 'Bob Tran', '00000000-0000-0000-0000-000000000001'),

  -- CS202 (OOP) → Alice Nguyen, Room A102, Mon period 3
  ('00000000-0000-0008-0000-000000000002', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000006', '00000000-0000-0001-0000-000000000001',
   '00000000-0000-0003-0000-000000000002', '00000000-0000-0006-0001-000000000003',
   false, 'Object-Oriented Programming', 'CS202', 'Alice Nguyen', '00000000-0000-0000-0000-000000000001'),

  -- MATH201 (Calc II) → Carol Le, Hall D301, Tue period 1
  ('00000000-0000-0008-0000-000000000003', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000008', '00000000-0000-0001-0000-000000000003',
   '00000000-0000-0003-0000-000000000007', '00000000-0000-0006-0002-000000000001',
   false, 'Calculus II', 'MATH201', 'Carol Le', '00000000-0000-0000-0000-000000000002'),

  -- CS301 (Algorithms) → Henry Ly, Room A103, Tue period 3
  ('00000000-0000-0008-0000-000000000004', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000011', '00000000-0000-0001-0000-000000000008',
   '00000000-0000-0003-0000-000000000003', '00000000-0000-0006-0002-000000000003',
   false, 'Algorithms', 'CS301', 'Henry Ly', '00000000-0000-0000-0000-000000000001'),

  -- CS302 (Databases) → Bob Tran, Lab B201, Wed period 2
  ('00000000-0000-0008-0000-000000000005', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000012', '00000000-0000-0001-0000-000000000002',
   '00000000-0000-0003-0000-000000000004', '00000000-0000-0006-0003-000000000002',
   false, 'Database Systems', 'CS302', 'Bob Tran', '00000000-0000-0000-0000-000000000001'),

  -- CS303 (Networks) → Grace Vu, Lab B202, Wed period 4
  ('00000000-0000-0008-0000-000000000006', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000013', '00000000-0000-0001-0000-000000000007',
   '00000000-0000-0003-0000-000000000005', '00000000-0000-0006-0003-000000000004',
   false, 'Computer Networks', 'CS303', 'Grace Vu', '00000000-0000-0000-0000-000000000001'),

  -- CS304 (OS) → Jack Mai, Room A101, Thu period 1
  ('00000000-0000-0008-0000-000000000007', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000014', '00000000-0000-0001-0000-000000000010',
   '00000000-0000-0003-0000-000000000001', '00000000-0000-0006-0004-000000000001',
   false, 'Operating Systems', 'CS304', 'Jack Mai', '00000000-0000-0000-0000-000000000001'),

  -- CS305 (Software Eng) → Alice Nguyen, Seminar C101, Thu period 3
  ('00000000-0000-0008-0000-000000000008', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000015', '00000000-0000-0001-0000-000000000001',
   '00000000-0000-0003-0000-000000000006', '00000000-0000-0006-0004-000000000003',
   false, 'Software Engineering', 'CS305', 'Alice Nguyen', '00000000-0000-0000-0000-000000000001'),

  -- MATH301 (Numerical) → Carol Le, Lab B201, Thu period 5
  ('00000000-0000-0008-0000-000000000009', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000016', '00000000-0000-0001-0000-000000000003',
   '00000000-0000-0003-0000-000000000004', '00000000-0000-0006-0004-000000000005',
   false, 'Numerical Methods', 'MATH301', 'Carol Le', '00000000-0000-0000-0000-000000000002'),

  -- PHYS201 (EM) → Frank Do, Room A103, Mon period 1
  ('00000000-0000-0008-0000-000000000010', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000010', '00000000-0000-0001-0000-000000000006',
   '00000000-0000-0003-0000-000000000003', '00000000-0000-0006-0001-000000000001',
   false, 'Electromagnetism', 'PHYS201', 'Frank Do', '00000000-0000-0000-0000-000000000003'),

  -- PHYS301 (Quantum) → Eve Hoang, Room A102, Fri period 2
  ('00000000-0000-0008-0000-000000000011', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000017', '00000000-0000-0001-0000-000000000005',
   '00000000-0000-0003-0000-000000000002', '00000000-0000-0006-0005-000000000002',
   false, 'Quantum Mechanics', 'PHYS301', 'Eve Hoang', '00000000-0000-0000-0000-000000000003'),

  -- CS401 (ML) → Iris Dao, Lab B201, Tue period 5
  ('00000000-0000-0008-0000-000000000012', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000018', '00000000-0000-0001-0000-000000000009',
   '00000000-0000-0003-0000-000000000004', '00000000-0000-0006-0002-000000000005',
   false, 'Machine Learning', 'CS401', 'Iris Dao', '00000000-0000-0000-0000-000000000001'),

  -- CS402 (Distributed) → Grace Vu, Room A101, Fri period 3
  ('00000000-0000-0008-0000-000000000013', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000019', '00000000-0000-0001-0000-000000000007',
   '00000000-0000-0003-0000-000000000001', '00000000-0000-0006-0005-000000000003',
   false, 'Distributed Systems', 'CS402', 'Grace Vu', '00000000-0000-0000-0000-000000000001'),

  -- CS403 (Compilers) → Iris Dao, Room A102, Wed period 5
  ('00000000-0000-0008-0000-000000000014', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000020', '00000000-0000-0001-0000-000000000009',
   '00000000-0000-0003-0000-000000000002', '00000000-0000-0006-0003-000000000005',
   false, 'Compiler Design', 'CS403', 'Iris Dao', '00000000-0000-0000-0000-000000000001'),

  -- CS404 (Security) → Grace Vu, Lab B202, Mon period 4
  ('00000000-0000-0008-0000-000000000015', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000021', '00000000-0000-0001-0000-000000000007',
   '00000000-0000-0003-0000-000000000005', '00000000-0000-0006-0001-000000000004',
   false, 'Computer Security', 'CS404', 'Grace Vu', '00000000-0000-0000-0000-000000000001'),

  -- CS405 (AI) → Iris Dao, Room A103, Fri period 4
  ('00000000-0000-0008-0000-000000000016', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000022', '00000000-0000-0001-0000-000000000009',
   '00000000-0000-0003-0000-000000000003', '00000000-0000-0006-0005-000000000004',
   false, 'Artificial Intelligence', 'CS405', 'Iris Dao', '00000000-0000-0000-0000-000000000001'),

  -- DB401 (Advanced DB) → Bob Tran, Lab B201, Fri period 1
  ('00000000-0000-0008-0000-000000000017', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000023', '00000000-0000-0001-0000-000000000002',
   '00000000-0000-0003-0000-000000000004', '00000000-0000-0006-0005-000000000001',
   false, 'Advanced Databases', 'DB401', 'Bob Tran', '00000000-0000-0000-0000-000000000001'),

  -- CS499 (Capstone) → Jack Mai, Seminar C101, Tue period 6 (manual override — moved by admin)
  ('00000000-0000-0008-0000-000000000018', '00000000-0000-0007-0000-000000000001',
   '00000000-0000-0002-0000-000000000024', '00000000-0000-0001-0000-000000000010',
   '00000000-0000-0003-0000-000000000006', '00000000-0000-0006-0002-000000000006',
   true, 'Capstone Project', 'CS499', 'Jack Mai', '00000000-0000-0000-0000-000000000001')
ON CONFLICT (id) DO NOTHING;

COMMIT;
