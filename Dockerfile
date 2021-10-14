FROM alpine
RUN apk add curl
WORKDIR /
COPY app /app
ENTRYPOINT ["/app"]
