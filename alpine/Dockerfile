ARG ALPINE_VERSION
FROM alpine:$ALPINE_VERSION
ARG GLIBC_VERSION

RUN apk add --no-cache curl libgcc
RUN curl -Ls https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub && \
    curl -LsO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-$GLIBC_VERSION.apk && \
    curl -LsO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-bin-$GLIBC_VERSION.apk && \
    apk add *.apk && rm *.apk
