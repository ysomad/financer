# https://github.com/pgroonga/docker

FROM postgres:16-alpine

RUN \
  apk --no-cache add \
    autoconf \
    automake \
    build-base \
    clang15-dev \
    gettext-dev \
    libtool \
    llvm15 \
    lz4-dev \
    msgpack-c-dev \
    zstd-dev

ENV PGROONGA_VERSION=3.1.6 \
    GROONGA_VERSION=13.1.1

COPY build/build_postgres.sh /build.sh
RUN chmod +x /build.sh
RUN \
  /build.sh ${PGROONGA_VERSION} ${GROONGA_VERSION} && \
  rm -f build.sh

FROM postgres:16-alpine

# copy thirdpart lib files
COPY --from=0 /usr/lib/libmsgpack-c.so.? /usr/lib/
COPY --from=0 /usr/lib/liblz4.so.? /usr/lib/
COPY --from=0 /usr/lib/libzstd.so.? /usr/lib/
# copy MeCab files
COPY --from=0 /usr/local/etc/mecabrc /usr/local/etc/
COPY --from=0 /usr/local/lib/libmecab.so.? /usr/local/lib/
COPY --from=0 /usr/local/lib/mecab/ /usr/local/lib/mecab/
# copy Groonga lib files
COPY --from=0 /usr/local/etc/groonga/ /usr/local/etc/groonga/
COPY --from=0 /usr/local/lib/groonga/ /usr/local/lib/groonga/
COPY --from=0 /usr/local/lib/libgroonga.so.? /usr/local/lib/
# copy PGroonga extension files
COPY --from=0 /usr/local/lib/postgresql/pgroonga*.so /usr/local/lib/postgresql/
COPY --from=0 /usr/local/share/postgresql/extension/pgroonga* /usr/local/share/postgresql/extension/
