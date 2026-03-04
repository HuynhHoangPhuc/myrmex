FROM golang:1.26-alpine AS builder
WORKDIR /workspace

COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY gen/ gen/
COPY services/module-student/ services/module-student/
COPY services/core/go.mod services/core/go.mod
COPY services/module-hr/go.mod services/module-hr/go.mod
COPY services/module-subject/go.mod services/module-subject/go.mod
COPY services/module-timetable/go.mod services/module-timetable/go.mod
COPY services/module-analytics/go.mod services/module-analytics/go.mod
COPY services/module-notification/go.mod services/module-notification/go.mod

WORKDIR /workspace/services/module-student
RUN go build -o /bin/module-student ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/module-student /bin/module-student
COPY --from=builder /workspace/services/module-student/config/ /config/

EXPOSE 50055
ENTRYPOINT ["/bin/module-student"]
