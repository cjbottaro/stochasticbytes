ARG ALPINE_VERSION
FROM alpine:${ALPINE_VERSION}

RUN apk add --no-cache openvpn
ADD shim.sh /

ENTRYPOINT ["/shim.sh"]
