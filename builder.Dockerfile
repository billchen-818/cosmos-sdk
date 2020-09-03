FROM golang:1.15.0-buster
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get --no-install-recommends -y install \
    pciutils build-essential git wget \
    lsb-release dpkg-dev curl bsdmainutils fakeroot
RUN useradd -ms /bin/bash -U builder
USER builder:builder
WORKDIR /sources
VOLUME [ "/artifacts" ]
VOLUME [ "/sources" ]
ENTRYPOINT [ "/sources/build.sh" ]
