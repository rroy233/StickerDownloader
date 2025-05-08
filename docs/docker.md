# 🐳 StickerDownloader Docker 部署教程（两种方式）

StickerDownloader 应用，支持以下两种场景：

1. **方式一**：使用 `docker-compose` 一键部署 Redis 和 App
2. **方式二**：使用外部 Redis，仅运行 App 容器（连接宿主机或云端 Redis）

---

## 🧰 准备前提

* 安装好 [Docker](https://docs.docker.com/get-docker/)
* 安装好 [Docker Compose](https://docs.docker.com/compose/)
* 克隆或下载本项目源代码
* 将 `config.example.yaml` 修改为实际配置并命名为 `config.yaml`

**镜像源选择**

| 镜像仓库                      | 镜像地址格式                                                     |
| ------------------------- |------------------------------------------------------------|
| Docker Hub                | `docker.io/rroy233/stickerdownloader`                      |
| GitHub Container Registry | `ghcr.io/rroy233/stickerdownloader`                       |
| 阿里云镜像仓库                   | `registry.cn-guangzhou.aliyuncs.com/rroy233/stickerdownloader` |

(本教程默认使用Docker Hub作为镜像仓库)

---

# ✅ 方式一：使用 Docker Compose 一键启动（含 Redis）

---

### 📄 1. 配置 config.yaml（Redis 使用内部服务）

```yaml
redis:
  server: "redis"          # 对应 docker-compose 的 redis 服务名
  port: "6379"
  password: ""
  tls: false
  db: 0
```

---

### 🧩 2. 修改默认的docker-compose.yml

```yaml
  app:
    image: rroy233/stickerdownloader:latest   # 可替换镜像
```

---

### ▶️ 3. 一键启动

```bash
docker-compose up --build -d
```

---

# ✅ 方式二：使用外部 Redis，仅运行 App 容器

---

## 📄 1. 配置 config.yaml（连接外部 Redis）

例如连接宿主机、云端或远程 Redis：

```yaml
redis:
  server: "host.docker.internal"  # 宿主机 Redis (推荐 macOS/Windows/Linux)
  # server: "172.17.0.1"          # Linux bridge 模式下宿主机 IP
  # server: "rds.aliyuncs.com"    # 也可为云 Redis 地址
  port: "6379"
  password: ""
  tls: false
  db: 0
```

---

## 🐳 2. 运行 App 镜像

### 启动容器

```bash
docker run -d \
  --name sticker_app \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/log:/app/log \
  -v $(pwd)/storage:/app/storage \
  rroy233/stickerdownloader:latest
```


# 🧼 清理

```bash
docker-compose down -v     # 方式一
docker rm -f sticker_app   # 方式二
```
