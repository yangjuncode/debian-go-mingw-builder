#! /bin/bash

go_version=1.24.1
build_no=1
go_version_build=${go_version}-${build_no}

docker build --progress=plain --build-arg go_version=${go_version} -t ghcr.io/yangjuncode/go:${go_version_build} -f Dockerfile.go .
docker tag ghcr.io/yangjuncode/go:${go_version_build} ghcr.io/yangjuncode/go:${go_version}

cmake_version=3.31.5

docker build --progress=plain --build-arg go_version_build=${go_version_build} --build-arg cmake_version=${cmake_version} -t ghcr.io/yangjuncode/go-mingw:${go_version_build} -f Dockerfile.go-mingw .
docker tag ghcr.io/yangjuncode/go-mingw:${go_version_build} ghcr.io/yangjuncode/go-mingw:${go_version}

#if has param -p, push to docker registry
if [[ $1 == "-p" ]]; then
    docker push ghcr.io/yangjuncode/go:${go_version_build}
    docker push ghcr.io/yangjuncode/go:${go_version}
    docker push ghcr.io/yangjuncode/go-mingw:${go_version_build}
    docker push ghcr.io/yangjuncode/go-mingw:${go_version}
fi
