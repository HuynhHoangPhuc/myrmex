FROM golang:1.26-alpine AS builder
WORKDIR /workspace

# Copy workspace files first (layer caching)
COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY gen/ gen/
COPY services/module-analytics/ services/module-analytics/
# go.mod stubs so go.work can resolve all workspace modules
COPY services/core/go.mod services/core/go.mod
COPY services/module-hr/go.mod services/module-hr/go.mod
COPY services/module-subject/go.mod services/module-subject/go.mod
COPY services/module-timetable/go.mod services/module-timetable/go.mod

WORKDIR /workspace/services/module-analytics
RUN go build -o /bin/module-analytics ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/module-analytics /bin/module-analytics
COPY --from=builder /workspace/services/module-analytics/config/ /config/

EXPOSE 8055
ENTRYPOINT ["/bin/module-analytics"]
