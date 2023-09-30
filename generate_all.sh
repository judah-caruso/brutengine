#!/usr/bin/env sh

# Generates json api and binding projects

go generate ./... && \
pushd bindings/brutengine_go/generate > /dev/null && \
go run . && \
popd > /dev/null
