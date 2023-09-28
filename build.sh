#!/usr/bin/env sh

go generate ./... && \
pushd bindings/brutengine_go/generate > /dev/null && \
go run . && \
popd > /dev/null

