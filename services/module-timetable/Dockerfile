FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copy workspace files for local module resolution
COPY go.work go.work.sum ./
COPY gen/go ./gen/go
COPY pkg ./pkg
COPY services/module-timetable ./services/module-timetable
# go.mod stubs so go.work can resolve all workspace modules
COPY services/core/go.mod services/core/go.mod
COPY services/module-hr/go.mod services/module-hr/go.mod
COPY services/module-subject/go.mod services/module-subject/go.mod
COPY services/module-analytics/go.mod services/module-analytics/go.mod

WORKDIR /workspace/services/module-timetable
RUN go build -o /bin/timetable-server ./cmd/server

# --- Runtime image ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/timetable-server /bin/timetable-server
COPY services/module-timetable/config /config

EXPOSE 50054

ENTRYPOINT ["/bin/timetable-server"]
