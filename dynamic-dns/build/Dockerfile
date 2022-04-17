ARG ALPINE_VERSION
ARG RUST_VERSION

FROM rust:${RUST_VERSION}-alpine${ALPINE_VERSION} AS build
RUN apk add build-base openssl-dev
ADD . /rust
WORKDIR /rust
# First flag fixes segfault from linking to openssl in Alpine.
# Second flag tries to statically link. Doesn't work for libgcc though.
ENV RUSTFLAGS "-C target-feature=-crt-static -C link-arg=-s"
RUN --mount=type=cache,target=/rust/target --mount=type=cache,target=/usr/local/cargo/registry cargo build --release
RUN --mount=type=cache,target=/rust/target cp /rust/target/release/dynamic-dns /dynamic-dns

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache libgcc
COPY --from=build /dynamic-dns /dynamic-dns
ENTRYPOINT [ "/dynamic-dns" ]