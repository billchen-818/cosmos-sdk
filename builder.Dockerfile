FROM golang:1.15.0-buster
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get --no-install-recommends -y install \
    pciutils build-essential git wget \
    lsb-release dpkg-dev curl bsdmainutils fakeroot
RUN useradd -ms /bin/bash -U builder
ARG PACKAGE
ENV PACKAGE ${PACKAGE:-cosmos-sdk}
USER builder:builder
WORKDIR /sources
VOLUME [ "/sources" ]
ENTRYPOINT [ "/sources/build.sh" ]
