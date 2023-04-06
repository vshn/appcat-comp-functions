FROM docker.io/library/alpine:3.17 as runtime

RUN \
	apk add --update --no-cache \
	bash \
	curl \
	ca-certificates \
	tzdata

ENTRYPOINT ["appcat-comp-functions"]
CMD ["--log-level", "1"]
COPY appcat-comp-functions /usr/bin/
