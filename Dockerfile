# ---- Build stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /omniroute ./cmd/omniroute

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates jq sqlite tzdata wget \
    && adduser -D -h /app omniroute

WORKDIR /app

COPY --from=builder /omniroute /app/omniroute
COPY --chmod=755 scripts/check-data-dir.sh /app/check-data-dir.sh

RUN mkdir -p /app/data && chown omniroute:omniroute /app/data

USER omniroute

EXPOSE 3456

ENV DATA_DIR=/app/data
ENV PORT=3456

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- --timeout=4 --tries=1 "http://127.0.0.1:${PORT}/health" \
      | jq -e 'type == "object" and .status == "ok" and .db == "ok"' >/dev/null

ENTRYPOINT ["/app/check-data-dir.sh"]
CMD ["/app/omniroute"]
