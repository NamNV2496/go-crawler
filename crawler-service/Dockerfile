# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git make curl
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/crawler main.go
FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl tzdata
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
WORKDIR /app
COPY --from=builder /app/bin/crawler /app/crawler
COPY --from=builder /app/.env.example* /app/ || true
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD pgrep crawler || exit 1
ENTRYPOINT ["/app/crawler"]
CMD ["crawler-worker"]
