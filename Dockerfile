FROM golang:1.23 as builder
LABEL authors="Roy"

WORKDIR /app
COPY . .

RUN go build -o app

FROM debian:bookworm-slim
# 安装FFmpeg与Redis CLI
RUN apt-get update && \
    apt-get install -y ffmpeg redis-tools && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app /app

VOLUME ["/app/storage", "/app/log"]

# 默认启动程序
CMD ["./app"]