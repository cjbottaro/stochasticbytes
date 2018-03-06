FROM alpine:3.7

RUN apk add --no-cache openvpn
ADD shim.sh /

ENTRYPOINT ["/shim.sh"]
