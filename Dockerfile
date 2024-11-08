FROM golang:1.22-alpine AS build

WORKDIR /webhook

COPY pkg/ .
RUN apk --no-cache add build-base gcc && \
    adduser -S 10001 golang && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o main

FROM alpine:3.20

COPY --from=build /webhook/main .
COPY --from=build /etc/passwd /etc/passwd

ENTRYPOINT [ "/main" ]