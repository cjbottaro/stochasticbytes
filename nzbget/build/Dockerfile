ARG ALPINE_VERSION
FROM alpine:${ALPINE_VERSION}

ARG NZBGET_VERSION
ARG NZBGET_RELEASE

ADD ${NZBGET_RELEASE} nzbget-bin-linux.run
RUN sh nzbget-bin-linux.run
WORKDIR /nzbget
ADD watchdog.sh shim.sh ./
ENTRYPOINT ["/bin/sh", "shim.sh"]
