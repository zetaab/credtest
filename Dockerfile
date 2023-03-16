FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk --no-cache add ca-certificates && \
    rm -rf /var/cache/apk/*

WORKDIR /code
USER 1001
COPY credtest .
ENTRYPOINT ["/code/credtest"]
