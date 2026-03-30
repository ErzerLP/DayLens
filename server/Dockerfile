# ===== 构建阶段 =====
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# 依赖缓存
COPY go.mod go.sum ./
RUN go mod download

# 编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/server \
    ./cmd/server

# ===== 运行阶段 =====
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/server /app/server
COPY --from=builder /build/migrations /app/migrations

# 数据目录
RUN mkdir -p /app/data/screenshots

EXPOSE 8080 8081

ENTRYPOINT ["/app/server"]
