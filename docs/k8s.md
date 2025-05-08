# ☸️ StickerDownloader Kubernetes 部署教程

本教程将引导你将 StickerDownloader 和 Redis 部署到 Kubernetes 集群中，支持配置注入、持久化、自动重启与更新。

---

## 📁 一、准备配置文件 config.yaml

### 1. 从 `config.example.yaml` 创建：

```bash
cp config.example.yaml config.yaml
```

### 2. 修改 Redis 地址为 Kubernetes 内部服务名：

```yaml
redis:
  server: "redis"       # Redis Service 名称，Cluster 内 DNS 可解析
  port: "6379"
  password: ""
  tls: false
  db: 0
```

### 3. 创建 Secret 以挂载配置（推荐包含敏感信息如 token）：

```bash
kubectl create secret generic app-config \
  --from-file=config.yaml=./config.yaml
```

> 📌 也可写为 YAML 文件形式或搭配 GitOps 工具（如 Kustomize、ArgoCD）

---

## 🔄 二、部署与管理

### 启动部署：

```bash
kubectl apply -f k8s/*
```

### 更新配置：

修改 `config.yaml` 后：

```bash
kubectl delete secret app-config
kubectl create secret generic app-config --from-file=config.yaml=./config.yaml
kubectl rollout restart deployment sticker-app
```

或使用 `kubectl patch` 替换 `configMap`/`secret` 内容。

---

## 🔁 三、更新镜像

### 1. 修改 image tag 并更新：

```yaml
image: ghcr.io/rroy233/stickerdownloader:v1.2.3
```

然后：

```bash
kubectl apply -f k8s/app-deployment.yaml
```

### 2. 或自动触发滚动更新：

```bash
kubectl rollout restart deployment sticker-app
```

---

## ✅ 四、验证运行状态

```bash
kubectl get pods
kubectl logs -f deploy/sticker-app
```