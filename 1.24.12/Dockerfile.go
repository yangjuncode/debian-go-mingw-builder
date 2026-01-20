FROM buildpack-deps:bullseye-scm AS build

ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION 1.24.12

RUN set -eux; \
	now="$(date '+%s')"; \
	arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; \
	url=; \
	case "$arch" in \
		'amd64') \
			url='https://dl.google.com/go/go1.24.12.linux-amd64.tar.gz'; \
			sha256='bddf8e653c82429aea7aec2520774e79925d4bb929fe20e67ecc00dd5af44c50'; \
			;; \
		'armhf') \
			url='https://dl.google.com/go/go1.24.12.linux-armv6l.tar.gz'; \
			sha256='c1a69349f2a511ecfc1344747549cbd6d69fae2f3a236b7dd952dcb1cff4d48c'; \
			;; \
		'arm64') \
			url='https://dl.google.com/go/go1.24.12.linux-arm64.tar.gz'; \
			sha256='4e02e2979e53b40f3666bba9f7e5ea0b99ea5156e0824b343fd054742c25498d'; \
			;; \
		'i386') \
			url='https://dl.google.com/go/go1.24.12.linux-386.tar.gz'; \
			sha256='51fe85c095908c992a63aa2126a4d274226da44af6b31ec4df6ee8cbb6acc497'; \
			;; \
		'mips64el') \
			url='https://dl.google.com/go/go1.24.12.linux-mips64le.tar.gz'; \
			sha256='85b384891cf6a4c265dcaf1e352ba7edaa952d0159abd03389c4b0cdcd950633'; \
			;; \
		'ppc64el') \
			url='https://dl.google.com/go/go1.24.12.linux-ppc64le.tar.gz'; \
			sha256='5db03a5e6318749ebbfd0e5c8ad6b6bd01b025cc8d878c5884991513de8e05a8'; \
			;; \
		'riscv64') \
			url='https://dl.google.com/go/go1.24.12.linux-riscv64.tar.gz'; \
			sha256='9d4340f8d8fbf0e89772aab1de79ce7a7d1f128173dcc8f9f0d07e009187b232'; \
			;; \
		's390x') \
			url='https://dl.google.com/go/go1.24.12.linux-s390x.tar.gz'; \
			sha256='f0fb0409aedce5c94ce38c85060f06ae6b3d8ecb6e0556ec0f016f4acb6f616e'; \
			;; \
		*) echo >&2 "error: unsupported architecture '$arch' (likely packaging update needed)"; exit 1 ;; \
	esac; \
	\
	wget -O go.tgz.asc "$url.asc"; \
	wget -O go.tgz "$url" --progress=dot:giga; \
	echo "$sha256 *go.tgz" | sha256sum -c -; \
	\
	GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
	gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'; \
	gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13'; \
	gpg --batch --verify go.tgz.asc go.tgz; \
	gpgconf --kill all; \
	rm -rf "$GNUPGHOME" go.tgz.asc; \
	\
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	SOURCE_DATE_EPOCH="$(stat -c '%Y' /usr/local/go)"; \
	export SOURCE_DATE_EPOCH; \
	touchy="$(date -d "@$SOURCE_DATE_EPOCH" '+%Y%m%d%H%M.%S')"; \
	date --date "@$SOURCE_DATE_EPOCH" --rfc-2822; \
	[ "$SOURCE_DATE_EPOCH" -lt "$now" ]; \
	\
	if [ "$arch" = 'armhf' ]; then \
		[ -s /usr/local/go/go.env ]; \
		before="$(go env GOARM)"; [ "$before" != '7' ]; \
		{ \
			echo; \
			echo '# https://github.com/docker-library/golang/issues/494'; \
			echo 'GOARM=7'; \
		} >> /usr/local/go/go.env; \
		after="$(go env GOARM)"; [ "$after" = '7' ]; \
		touch -t "$touchy" /usr/local/go/go.env /usr/local/go; \
	fi; \
	\
	mkdir /target /target/usr /target/usr/local; \
	mv -vT /usr/local/go /target/usr/local/go; \
	ln -svfT /target/usr/local/go /usr/local/go; \
	touch -t "$touchy" /target/usr/local /target/usr /target; \
	\
	go version; \
	epoch="$(stat -c '%Y' /target/usr/local/go)"; \
	[ "$SOURCE_DATE_EPOCH" = "$epoch" ]; \
	find /target -newer /target/usr/local/go -exec sh -c 'ls -ld "$@" && exit "$#"' -- '{}' +

FROM buildpack-deps:bullseye-scm

LABEL org.opencontainers.image.source=https://github.com/yangjuncode/debian-go-mingw-builder

USER root

RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
		binutils-gold \
		patch \
		autoconf \
		libtool \
		automake \
	; \
	rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.24.12
ENV GOTOOLCHAIN=local

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

COPY --from=build --link /target/ /

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 1777 "$GOPATH"

WORKDIR $GOPATH

COPY patch/*.patch /usr/local/go/

RUN cd /usr/local/go \
	&& for patch_file in *.patch; do \
		patch --verbose -p1 < "$patch_file"; \
	done \
	&& rm -rf *.patch
