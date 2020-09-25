FROM golang:1.14-alpine
WORKDIR /omniparser
COPY . .
RUN go build -o cli/op cli/op.go
RUN cli/op --help
CMD cli/op server
