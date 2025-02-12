ARG go_version=1.22.11
FROM golang:${go_version}-bullseye

LABEL org.opencontainers.image.source=https://github.com/yangjuncode/debian-go-mingw-builder

USER root

RUN set -o xtrace \
    && apt-get clean \
    && apt-get update \
    && apt-get -y --allow-unauthenticated install patch

COPY  patch/*.patch /usr/local/go/

RUN cd /usr/local/go \
    && for patch_file in *.patch; do \
    patch --verbose -p1 < "/usr/local/go/$patch_file"; \
    done \
    && rm -rf /usr/local/go/*.patch