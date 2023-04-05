FROM golang:1.20-alpine as Build

ARG INSTANCE=""

WORKDIR /app

COPY . ./
RUN echo $(ls -a) && go mod download

RUN cd "cmd/$INSTANCE" && CGO_ENABLED=0 go build -o functionio .

FROM docker.io/library/alpine:3.17 as runtime

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
