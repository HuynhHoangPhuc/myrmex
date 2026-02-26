FROM golang:1.26-alpine AS builder
WORKDIR /workspace

# Copy workspace files first (layer caching)
COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY gen/ gen/
COPY services/core/ services/core/
# go.mod stubs so go.work can resolve all workspace modules
COPY services/module-hr/go.mod services/module-hr/go.mod
COPY services/module-subject/go.mod services/module-subject/go.mod
COPY services/module-timetable/go.mod services/module-timetable/go.mod
COPY services/module-analytics/go.mod services/module-analytics/go.mod

WORKDIR /workspace/services/core
RUN go build -o /bin/core ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/core /bin/core
COPY --from=builder /workspace/services/core/config/ /config/

EXPOSE 8080 50051
ENTRYPOINT ["/bin/core"]
