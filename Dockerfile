FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk --no-cache add ca-certificates && \
    rm -rf /var/cache/apk/*

USER 1001
WORKDIR /code
RUN mkdir /code/.aws && chmod -R 777 /code
COPY credtest .
ENV HOME=/code
ENTRYPOINT ["/code/credtest"]
