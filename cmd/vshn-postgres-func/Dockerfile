FROM docker.io/library/alpine:3.17 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["vshn-postgres-func"]
CMD ["--log-level", "1"]
COPY vshn-postgres-func /usr/bin/

USER 65536:0
