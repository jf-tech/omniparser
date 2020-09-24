FROM golang:1.14-alpine
WORKDIR /omniparser
COPY . .
RUN go build -o omniparser/cli/op omniparser/cli/op.go
RUN omniparser/cli/op --help
CMD omniparser/cli/op server
