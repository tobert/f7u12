FROM       busybox
MAINTAINER Al Tobey <atobey@datastax.com>

RUN mkdir /public
COPY f7u12 /
COPY public /public
EXPOSE 8080
USER 1337
ENTRYPOINT ["/f7u12", "-addr", ":8080"]
