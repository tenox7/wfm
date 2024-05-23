FROM scratch
ARG TARGETARCH
ADD wfm-${TARGETARCH}-linux /wfm
ENTRYPOINT ["/wfm"]
LABEL maintainer="as@tenoware.com"