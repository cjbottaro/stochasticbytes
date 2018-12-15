FROM alpine:3.7
ARG NZBGET_VERSION
ADD https://github.com/nzbget/nzbget/releases/download/v$NZBGET_VERSION/nzbget-$NZBGET_VERSION-bin-linux.run .
RUN sh nzbget-$NZBGET_VERSION-bin-linux.run
WORKDIR /nzbget
ADD watchdog.sh shim.sh ./
ENTRYPOINT ["/bin/sh", "shim.sh"]
