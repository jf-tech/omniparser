FROM golang:1.14-alpine
WORKDIR /omniparser
COPY . .
RUN mkdir bin
RUN go build -o bin/op omniparser/cli/op.go
COPY ./omniparser/cli/cmd/web bin/cmd/web
RUN bin/op --help
CMD bin/op server
