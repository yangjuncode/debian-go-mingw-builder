#! /bin/bash

go_version=1.21.13-1

docker build --progress=plain -t ghcr.io/yangjuncode/go:${go_version} -f Dockerfile.go .
docker tag ghcr.io/yangjuncode/go:${go_version} ghcr.io/yangjuncode/go:1.21.13

docker build --progress=plain --build-arg go_version=${go_version} -t ghcr.io/yangjuncode/go-mingw:${go_version} -f Dockerfile.go-mingw .
docker tag ghcr.io/yangjuncode/go-mingw:${go_version} ghcr.io/yangjuncode/go-mingw:1.21.13

#if has param -p, push to docker registry
if [[ $1 == "-p" ]]; then
    docker push ghcr.io/yangjuncode/go:${go_version}
    docker push ghcr.io/yangjuncode/go:1.21.13
    docker push ghcr.io/yangjuncode/go-mingw:${go_version}
    docker push ghcr.io/yangjuncode/go-mingw:1.21.13
fi
