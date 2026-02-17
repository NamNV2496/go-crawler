# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/scheduler main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/scheduler /app/scheduler
COPY --from=builder /app/.env.example* /app/
ENTRYPOINT ["/app/scheduler"]
CMD ["scheduler"]
