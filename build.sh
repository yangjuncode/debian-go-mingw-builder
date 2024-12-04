#! /bin/bash

go_version=1.21.13

docker build --progress=plain -t ghcr.io/yangjuncode/go-mingw:${go_version}  .
