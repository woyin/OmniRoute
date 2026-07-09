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

RUN apk add --no-cache ca-certificates sqlite-libs tzdata \
    && adduser -D -h /app omniroute

WORKDIR /app

COPY --from=builder /omniroute /app/omniroute

RUN mkdir -p /app/data && chown omniroute:omniroute /app/data

USER omniroute

EXPOSE 3456

ENV DATA_DIR=/app/data
ENV PORT=3456

ENTRYPOINT ["/app/omniroute"]
