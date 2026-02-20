FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copy go workspace files
COPY go.work go.work.sum* ./
COPY pkg/ pkg/
COPY gen/ gen/
COPY services/module-subject/ services/module-subject/

WORKDIR /workspace/services/module-subject
RUN go build -o /bin/module-subject ./cmd/server/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/module-subject /bin/module-subject
COPY --from=builder /workspace/services/module-subject/config/ /config/

EXPOSE 50053

ENTRYPOINT ["/bin/module-subject"]
