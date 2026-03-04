FROM golang:1.26-alpine AS builder
WORKDIR /workspace

# Copy workspace files first (layer caching)
COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY gen/ gen/
COPY services/module-notification/ services/module-notification/
# go.mod stubs so go.work can resolve all workspace modules
COPY services/core/go.mod services/core/go.mod
COPY services/module-hr/go.mod services/module-hr/go.mod
COPY services/module-subject/go.mod services/module-subject/go.mod
COPY services/module-timetable/go.mod services/module-timetable/go.mod
COPY services/module-student/go.mod services/module-student/go.mod
COPY services/module-analytics/go.mod services/module-analytics/go.mod

WORKDIR /workspace/services/module-notification
RUN go build -o /bin/module-notification ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/module-notification /bin/module-notification
COPY --from=builder /workspace/services/module-notification/config/ /config/

EXPOSE 8056
ENTRYPOINT ["/bin/module-notification"]
