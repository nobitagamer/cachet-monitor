FROM golang:1.12-alpine

# install build tools
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# build
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
WORKDIR /go/src/app/cli
RUN go build 

# start
CMD ["cli", "start"]