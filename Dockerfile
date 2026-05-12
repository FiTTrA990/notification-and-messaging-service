# ─────────────────────────────────────────────────────────────
# Stage 1 – Build
# ─────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

ARG BUILD_NUMBER=local
ARG GIT_COMMIT=unknown

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.BuildNumber=${BUILD_NUMBER} -X main.GitCommit=${GIT_COMMIT}" \
    -o notification-service \
    ./cmd/server

# ─────────────────────────────────────────────────────────────
# Stage 2 – Runtime (minimal image)
# ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime \
    && echo "Asia/Jakarta" > /etc/timezone

WORKDIR /app
COPY --from=builder /app/notification-service .

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./notification-service"]
