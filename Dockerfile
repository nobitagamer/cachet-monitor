# build image
FROM golang:1.12-alpine3.9 AS build-env

# install build tools
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# build
WORKDIR /go/src/app
COPY . .
RUN cd cli
RUN go build -mod=vendor

# distribution image
FROM alpine:3.9

# add CAs
RUN apk --no-cache add ca-certificates

WORKDIR /go/bin
COPY --from=build-env /go/src/app/cli ./cli

# start
CMD ["./cli", "start"]