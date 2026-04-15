FROM golang:1.23 as builder
LABEL authors="Roy"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y ffmpeg redis-tools ca-certificates tzdata && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/languages ./languages

VOLUME ["/app/storage", "/app/log"]

# 默认启动程序
CMD ["./app"]