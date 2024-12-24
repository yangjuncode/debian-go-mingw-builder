FROM golang:1.21-bullseye

#ref from https://github.com/yangjuncode/flutter-android-go-builder
LABEL org.opencontainers.image.source=https://github.com/yangjuncode/debian-go-mingw-builder/1.21

USER root

RUN set -o xtrace \
   && apt-get clean \
    && apt-get update \
    && apt-get -y --allow-unauthenticated install patch

COPY  patch/*.patch /usr/local/go/

RUN cd /usr/local/go \
    && patch --verbose -p1 < /usr/local/go/00-apply_https___go_dev_cl_600296.patch \
    && patch --verbose -p1 < /usr/local/go/01-Revert_crypto_rand,runtime__switch_RtlGenRandom_for_ProcessPrng.patch \
    && patch --verbose -p1 < /usr/local/go/02-fix__apply_https___go_dev_cl_600296.patch \
    && patch --verbose -p1 < /usr/local/go/03-runtime__reserve_4kB_for_system_stack_on_windows-386.patch \
    && rm -rf /usr/local/go/*.patch

