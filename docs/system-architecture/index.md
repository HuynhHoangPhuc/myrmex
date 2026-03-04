# System Architecture

## Overview

Myrmex is a microservice architecture with modular services communicating via gRPC and event streaming through NATS JetStream. The HTTP gateway (Core) proxies requests to gRPC services; the AI chat agent orchestrates operations via a dynamic tool registry. Audit logging captures mutations asynchronously via NATS → PostgreSQL for compliance and forensics.

## Architecture Components

- **API Gateway** (`Core` service): HTTP entry point, auth, module routing, WebSocket chat
- **Microservices**: 6 modules (HR, Subject, Timetable, Student, Analytics, Notification) via gRPC or HTTP
- **Event Bus**: NATS JetStream for event streaming, audit logs, and inter-service communication
- **Database**: PostgreSQL with schema-per-module isolation + shared audit logs
- **Cache**: Redis for shared caching abstractions (preferences, prerequisite validation)
- **Frontend**: React 19 with TanStack Router/Query, real-time notifications via WebSocket

## Key Pipelines

### Audit Logging Pipeline

Async fire-and-forget NATS pipeline with persistent storage:

1. **Middleware Capture** (audit_middleware.go):
   - Post-handler Gin middleware intercepts responses
   - Derives action from HTTP method: POST→Create, PATCH→Update, DELETE→Delete, GET→Read (skipped)
   - Publishes to NATS subject `AUDIT.logs` with user_id, resource_type, action, old/new values

2. **Async Consumer** (audit_consumer.go):
   - Durable JetStream consumer listening on `AUDIT.logs` stream
   - Writes to `core.audit_logs` partitioned table (12 monthly partitions 2026-03→2027-02)
   - Acknowledges on success; nack on error (exponential backoff retry)

3. **Storage Layer**:
   - Partitioned table `core.audit_logs` with monthly partitions
   - Columns: id, user_id, resource_type, action, old_value, new_value, timestamp
   - Indexes: BRIN (timestamp range), B-tree (user_id, resource_type)

4. **Query API**:
   - Endpoint: GET `/api/audit-logs` (admin/super_admin only)
   - Filters: user_id, resource_type, action, date range
   - Pagination: limit, offset (default 100 records per page)

5. **Frontend**:
   - Route: `/admin/audit-logs` with table UI, row expansion, filters
   - Sortable columns: User, Resource Type, Action, Timestamp
   - JSON diff rendering for old/new value comparison

### Notifications System (Phase 4.4 - COMPLETE)

Async NATS pipeline + Module-Notification microservice for email + in-app delivery:

**Architecture**:
- Domain events published to NATS subjects: `schedule.*`, `enrollment.*`, `grade.*`, `new_announcement`, `role_updated`, `user.deleted`
- Module-Notification consumes events, routes to email/in-app based on user preferences
- Email dispatch via SMTP (go-mail) with MJML templates
- In-app notifications pushed via WebSocket relay in Core service
- Exponential backoff retry (1h, 4h, 12h, 24h, 48h) for failed emails

**Preference Matrix**:
- 12 event types × 2 channels (email + in-app) per user
- Users can opt-in/opt-out per channel and event type
- Defaults: All channels enabled; user can selectively disable

**Storage**:
- `notification.notifications`: user_id, event_type, payload, is_read, created_at, read_at
- `notification.preferences`: user_id, event_type, channel, enabled
- `notification.email_queue`: to_email, subject, body, status, retry_count, next_retry_at

**API Endpoints** (Module-Notification on port 8056):
- GET `/notifications` (paginated list, filters by read status, type)
- POST `/notifications/:id/mark-read` (mark single notification as read)
- PATCH `/preferences` (update user notification preferences, bulk matrix update)
- GET `/preferences` (fetch current user's preference matrix)
- POST `/announcement` (admin-only, broadcast announcement to all users)
- HTTP proxy routes in Core gateway at `/api/notifications/*`

**Frontend Integration**:
- **Notifications Page** (`/notifications`): Paginated list, filters, delete actions
- **Preferences Page** (`/notifications/preferences`): 12×2 matrix UI with toggle switches
- **Notification Toast**: Real-time WS-pushed notifications with auto-dismiss
- **Sidebar Nav**: Notifications item with unread count badge
- **Hooks**: `use-notifications.ts` (list + mutations), `use-notification-ws.ts` (WebSocket stream)

**Configuration**:
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD` (env vars or config/local.yaml)
- `NOTIFICATION_FROM_EMAIL` (sender address for notification emails)
- Graceful degradation: If SMTP not configured, email notifications silently skipped

## Contents

- [Services Reference](./services.md) — Service topology, entity definitions, key operations
- [Data Flow](./data-flow.md) — Request/response patterns, event flows, chat agent workflow
- [Database Schema](./database-schema.md) — Logical schema definition for all modules
- [API Endpoints](./api-endpoints.md) — Complete API reference (HTTP + gRPC)

## System Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Client Layer                                   │
│  ┌──────────────────────┐  ┌──────────────────────────────────────┐    │
│  │   React Frontend     │  │   External API (if needed)           │    │
│  │   (localhost:3000)   │  │                                      │    │
│  └──────────┬───────────┘  └──────────────────────────────────────┘    │
└─────────────┼──────────────────────────────────────────────────────────┘
              │
              │ HTTP/WebSocket
              │
┌─────────────▼──────────────────────────────────────────────────────────┐
│                      API Gateway Layer                                  │
│                    Core Service (port 8080)                             │
│  ┌────────────────────────────────────────────────────────────────┐   │
│  │  Gin HTTP Router                                               │   │
│  │  ├─ POST /api/auth/register        → AuthService gRPC         │   │
│  │  ├─ POST /api/auth/login           → AuthService gRPC         │   │
│  │  ├─ POST /api/auth/refresh         → AuthService gRPC         │   │
│  │  ├─ GET  /api/auth/me              → UserService gRPC         │   │
│  │  ├─ GET  /api/dashboard/stats      → Dashboard (aggregate)    │   │
│  │  ├─ ANY  /api/hr/*                 → Module-HR gRPC (proxy)   │   │
│  │  ├─ ANY  /api/subjects/*           → Module-Subject gRPC      │   │
│  │  ├─ ANY  /api/timetable/*          → Module-Timetable gRPC    │   │
│  │  ├─ ANY  /api/students/*           → Module-Student gRPC      │   │
│  │  ├─ ANY  /api/analytics/*          → Module-Analytics HTTP    │   │
│  │  ├─ ANY  /api/notifications/*      → Module-Notification HTTP │   │
│  │  └─ WebSocket /ws/chat?token=X    → ChatGateway (Streaming)   │   │
│  │                                                                │   │
│  │  Middleware:                                                   │   │
│  │  ├─ CORS (origin validation)                                  │   │
│  │  ├─ Auth (JWT extraction + validation)                        │   │
│  │  ├─ Audit Logging (capture mutations → NATS)                  │   │
│  │  ├─ Rate Limiting (configurable per endpoint)                 │   │
│  │  └─ Request/Response logging                                  │   │
│  └────────────────────────────────────────────────────────────────┘   │
└─────────────┬──────────────────────────────────────────────────────────┘
              │          │           │              │             │
              │ gRPC     │ gRPC      │ gRPC         │ HTTP        │ HTTP
              │          │           │              │             │
     ┌────────▼──┐ ┌────▼──────┐ ┌─▼─────────┐ ┌─▼──────────┐ ┌▼──────────┐ ┌──────────────┐
     │Module-HR  │ │Module-    │ │ Module-   │ │ Module-    │ │Module-    │ │Module-       │
     │(50052)    │ │Subject    │ │Timetable  │ │ Student    │ │Analytics  │ │Notification │
     │           │ │(50053)    │ │ (50054)   │ │ (50055)    │ │(8055)     │ │(8056)        │
     └─────┬─────┘ └────┬──────┘ └─────┬─────┘ └──┬─────────┘ └────┬─────┘ └──────┬───────┘
           │            │              │           │               │             │
           └────────────┼──────────────┼───────────┼───────────────┼─────────────┘
                        │
                  NATS JetStream (Event Bus, port 4222)
                        │
        ┌───────────────┴──────────────────────────┐
        │                                          │
   ┌────▼─────┐  ┌──────────┐  ┌──────────┐  ┌───▼──────┐
   │ Event    │  │ Async    │  │Analytics │  │Notification│
   │ Store    │  │ Consumers│  │Consumer  │  │ Consumer   │
   │ (Append) │  │ (Listen) │  │(ETL)     │  │(Email/IA)  │
   └──────────┘  └──────────┘  └──────────┘  └────────────┘
        │
┌───────▼─────────────────────────────────────────────┐
│     PostgreSQL (Shared Database)                    │
│     localhost:5432 (myrmex / myrmex_dev)            │
│                                                     │
│ Schemas: core, hr, subject, timetable, student,    │
│          notification, analytics                   │
└─────────────────────────────────────────────────────┘
```

## Key Design Patterns

- **Schema-per-Module**: Reduces blast radius if one module is compromised
- **Event-Driven**: NATS JetStream for decoupled, asynchronous communication
- **CQRS**: Separate command (write) + query (read) handlers per service
- **Domain-Driven Design**: Entity aggregates with domain services and repositories
- **Circuit Breaker** (future): Wrap gRPC calls with retry + circuit breaker
- **Graceful Degradation**: Optional features (NATS, SMTP, OAuth) are no-op if not configured

## Performance Targets (MVP)

- API Latency (p95): ~300ms (excl. CSP)
- CSP Solver: ~20s for 100-subject semesters (30s timeout)
- DB Query: ~50ms (sqlc optimized)
- gRPC Call: ~100ms (local network)
- WebSocket Latency: ~50ms
- Frontend Bundle: ~80KB gzipped (Vite tree-shaking)
- Memory per Service: ~100MB (Go efficiency)
