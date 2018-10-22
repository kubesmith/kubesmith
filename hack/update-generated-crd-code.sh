#!/bin/bash

# redo this script to use docker so things are sandboxed (as opposed to just trampling all over host)

# setup a few helper variables
ROOT="github.com/kubesmith/kubesmith"
VERSION="kubernetes-1.11.0"

# remove the local versions of the generated code
rm "${GOPATH}/src/${ROOT}/pkg/apis/kubesmith/v1/zz_generated.deepcopy.go"
rm -rf "${GOPATH}/src/${ROOT}/pkg/generated"

# retrieve the code-generator scripts and bins
rm -rf "${GOPATH}/src/k8s.io"
mkdir -p "${GOPATH}/src/k8s.io"
cd "${GOPATH}/src/k8s.io/"
git clone -b ${VERSION} https://github.com/kubernetes/code-generator
git clone -b ${VERSION} https://github.com/kubernetes/apimachinery
cd code-generator

# run the code-generator entrypoint script
./generate-groups.sh all "${ROOT}/pkg/generated" "${ROOT}/pkg/apis" "kubesmith:v1"
