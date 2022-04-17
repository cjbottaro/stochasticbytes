ARG ALPINE_VERSION
FROM alpine:${ALPINE_VERSION} AS build
ARG NZBGET_RELEASE
ADD ${NZBGET_RELEASE} nzbget-bin-linux.run
RUN sh nzbget-bin-linux.run

FROM alpine:${ALPINE_VERSION}
COPY --from=build /nzbget /nzbget
WORKDIR /nzbget
ADD watchdog.sh shim.sh ./
ENTRYPOINT ["/bin/sh", "shim.sh"]
