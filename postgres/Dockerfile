ARG POSTGRES_VERSION

FROM postgres:${POSTGRES_VERSION} AS builder

RUN apk add --no-cache build-base git clang15 llvm15
RUN git clone --branch v0.4.4 https://github.com/pgvector/pgvector.git
WORKDIR /pgvector
RUN make
RUN make install

FROM postgres:${POSTGRES_VERSION}

COPY --from=builder /usr/local/lib/postgresql/ /usr/local/lib/postgresql/
COPY --from=builder /usr/local/share/postgresql/extension /usr/local/share/postgresql/extension