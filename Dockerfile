FROM golang:1.20-alpine3.17 AS build-env

RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
WORKDIR /go/src/github.com/devtron-labs/lens
ADD . /go/src/github.com/devtron-labs/lens
RUN go install github.com/google/wire/cmd/wire@latest
RUN GOOS=linux make

FROM alpine:3.17
RUN apk add --no-cache ca-certificates
COPY --from=build-env  /go/src/github.com/devtron-labs/lens/lens .
COPY --from=build-env  /go/src/github.com/devtron-labs/lens/scripts/ .
RUN adduser -D devtron
RUN chown -R devtron:devtron ./lens
USER devtron
CMD ["./lens"]
