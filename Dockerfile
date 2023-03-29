FROM docker.io/library/alpine:3.17 as runtime

RUN \
	apk add --update --no-cache \
	bash \
	curl \
	ca-certificates \
	tzdata

COPY --from=Build /app/cmd/$INSTANCE/functionio /usr/bin/

ENTRYPOINT ["appcat-comp-functions"]
CMD ["--log-level", "1"]
COPY appcat-comp-functions /usr/bin/
