#!/bin/bash
set -e

IMAGE_NAME="rroy233/stickerdownloader"

TAG=${1:-latest}

PLATFORMS="linux/amd64"

docker buildx build . \
  --platform ${PLATFORMS} \
  --tag ${IMAGE_NAME}:${TAG} \
  --output type=docker \
  --build-arg BUILDKIT_INLINE_CACHE=1

echo "✅ 镜像构建完成：${IMAGE_NAME}:${TAG}"
