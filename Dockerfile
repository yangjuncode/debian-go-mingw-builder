FROM golang:1.20-bullseye

#ref from https://github.com/yangjuncode/flutter-android-go-builder
LABEL org.opencontainers.image.source=https://github.com/yangjuncode/debian-go-mingw-builder

USER root

RUN set -o xtrace \
    && sed -i s@/deb.debian.org/@/mirrors.tuna.tsinghua.edu.cn/@g /etc/apt/sources.list \
   && apt-get clean \
    && apt-get update \
    && apt-get -y --allow-unauthenticated install automake cmake gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64 \
    && rm -rf /var/lib/apt/lists/*

