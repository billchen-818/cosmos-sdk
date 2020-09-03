#!/bin/bash

set -ex

declare -A OS_ARCHS
OS_ARCHS[linux]='amd64 arm64'
OS_ARCHS[darwin]='amd64'
OS_ARCHS[windows]='amd64'
export GO111MODULE=on

TEMPDIR="$(mktemp -d)"

# Make release tarball
PACKAGE=cosmos-sdk
VERSION=$(git describe --tags | sed 's/^v//')
COMMIT=$(git log -1 --format='%H')
DISTNAME=${PACKAGE}-${VERSION}
SOURCEDIST=${TEMPDIR}/${DISTNAME}.tar.gz
git archive --format tar.gz --prefix ${DISTNAME}/ -o ${SOURCEDIST} HEAD

pushd ${TEMPDIR}
go env GOPATH ${TEMPDIR}/go

# Correct tar file order
#mkdir -p temp
#pushd temp
tar xf ${SOURCEDIST}
rm ${SOURCEDIST}
find ${PACKAGE}-* | sort | tar --no-recursion --mode='u+rw,go+r-w,a+X' --owner=0 --group=0 -c -T - | gzip -9n > $SOURCEDIST
popd

# Prepare GOPATH and install deps
distsrc=${TEMPDIR}/buildsources
mkdir -p ${distsrc}
pushd ${distsrc}
tar --strip-components=1 -xf $SOURCEDIST
go mod download
popd

OUTDIR=/artifacts
rm -rfv ${OUTDIR}/*
# Extract release tarball and build
for os in "${!OS_ARCHS[@]}"; do
    exe_file_extension=''
    if [ ${os} = windows ]; then
        exe_file_extension='.exe'
    fi
    for arch in ${OS_ARCHS[$os]} ; do
        # Build gaia tool suite
        pushd ${distsrc}
        GOOS="${os}" GOARCH="${arch}" GOROOT_FINAL="$(go env GOROOT)" \
        make build-simd BUILDDIR=${OUTDIR}/ LDFLAGS=-buildid=${VERSION} VERSION=alessio COMMIT=${COMMIT}
        mv ${OUTDIR}/simd${exe_file_extension} ${OUTDIR}/${DISTNAME}-${os}-${arch}-simd${exe_file_extension}
        popd # ${distsrc}
        echo Build finished for GOOS="${os}" GOARCH="${arch}"
        ls ${OUTDIR}/
    done
    unset exe_file_extension
done

cd ${OUTDIR} && sha256sum ./*

