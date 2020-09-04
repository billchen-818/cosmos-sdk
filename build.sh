#!/bin/bash

set -uex

declare -A OS_ARCHS
if [[ "${TARGET_OS}" =~ linux ]]; then
    OS_ARCHS[linux]='amd64 arm64'
fi
if [[ "${TARGET_OS}" =~ darwin ]]; then
    OS_ARCHS[darwin]='amd64'
fi
if [[ "${TARGET_OS}" =~ windows ]]; then
    OS_ARCHS[windows]='amd64'
fi

TEMPDIR="$(mktemp -d)"

# Make release tarball
PACKAGE=${APP}
VERSION=$(git describe --tags | sed 's/^v//')
COMMIT=$(git log -1 --format='%H')
DISTNAME=${PACKAGE}-${VERSION}
SOURCEDIST=${TEMPDIR}/${DISTNAME}.tar.gz
git archive --format tar.gz --prefix ${DISTNAME}/ -o ${SOURCEDIST} HEAD

pushd ${TEMPDIR}

# Correct tar file order
tar xf ${SOURCEDIST}
rm ${SOURCEDIST}
find ${PACKAGE}-* | sort | tar --no-recursion --mode='u+rw,go+r-w,a+X' --owner=0 --group=0 -c -T - | gzip -9n > $SOURCEDIST
popd

# Extract release tarball and install deps
distsrc=${TEMPDIR}/buildsources
mkdir -p ${distsrc}
pushd ${distsrc}
tar --strip-components=1 -xf $SOURCEDIST
go mod download
popd

OUTDIR=$HOME/artifacts
BUILDDIR=${distsrc}/build
rm -rfv ${OUTDIR}/*
mkdir -p ${OUTDIR}/
for os in "${!OS_ARCHS[@]}"; do
    exe_file_extension=''
    if [ ${os} = windows ]; then
        exe_file_extension='.exe'
    fi
    for arch in ${OS_ARCHS[$os]} ; do
        # Build gaia tool suite
        pushd ${distsrc}
        GOOS="${os}" GOARCH="${arch}" GOROOT_FINAL="$(go env GOROOT)" \
        make build-simd \
            BUILDDIR=${BUILDDIR}/ \
            LDFLAGS=-buildid=${VERSION} \
            VERSION=${VERSION} \
            COMMIT=${COMMIT}
        mv ${BUILDDIR}/simd${exe_file_extension} ${OUTDIR}/${DISTNAME}-${os}-${arch}-simd${exe_file_extension}
        popd # ${distsrc}
        echo Build finished for GOOS="${os}" GOARCH="${arch}"
    done
    unset exe_file_extension
done

cd ${OUTDIR} && sha256sum ./*

