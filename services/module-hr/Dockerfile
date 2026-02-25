FROM golang:1.26-alpine AS builder
WORKDIR /workspace

# Copy workspace files first (layer caching)
COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY gen/ gen/
COPY services/module-hr/ services/module-hr/
# go.mod stubs so go.work can resolve all workspace modules
COPY services/core/go.mod services/core/go.mod
COPY services/module-subject/go.mod services/module-subject/go.mod
COPY services/module-timetable/go.mod services/module-timetable/go.mod

WORKDIR /workspace/services/module-hr
RUN go build -o /bin/module-hr ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/module-hr /bin/module-hr
COPY --from=builder /workspace/services/module-hr/config/ /config/

EXPOSE 50052
ENTRYPOINT ["/bin/module-hr"]
