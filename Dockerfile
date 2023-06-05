FROM golang:1.17.7-alpine3.15 AS build-env

RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
RUN go get github.com/google/wire/cmd/wire
WORKDIR /go/src/github.com/devtron-labs/lens
ADD . /go/src/github.com/devtron-labs/lens
RUN GOOS=linux make

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build-env  /go/src/github.com/devtron-labs/lens/lens .
RUN adduser -D devtron
RUN chown -R devtron:devtron ./lens
USER devtron
CMD ["./lens"]
