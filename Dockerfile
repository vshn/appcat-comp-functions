FROM golang:1.20-alpine as Build

ARG INSTANCE=""

WORKDIR /app

COPY . ./
RUN go mod download

RUN cd "cmd/$INSTANCE" && CGO_ENABLED=0 go build -o functionio .

FROM docker.io/library/alpine:3.17 as runtime

LABEL org.opencontainers.image.source=https://github.com/vshn/appcat-comp-functions
LABEL org.opencontainers.image.description="This repository has crossplane composition functions for appcat services"
LABEL org.opencontainers.image.licenses=BSD-3-Clause

ARG INSTANCE=""

RUN \
  apk add --update --no-cache \
  bash \
  curl \
  ca-certificates \
  tzdata

ENTRYPOINT ["functionio"]
CMD ["--log-level", "1"]
COPY --from=Build /app/cmd/$INSTANCE/functionio /usr/bin/
