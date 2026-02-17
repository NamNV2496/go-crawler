# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/crawler main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl tzdata
WORKDIR /app
COPY --from=builder /app/bin/crawler /app/crawler
COPY --from=builder /app/.env.example* /app/
ENTRYPOINT ["/app/crawler"]
CMD ["crawler-worker"]
