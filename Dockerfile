FROM golang:1.14
WORKDIR /omniparser
COPY . .
RUN go build -ldflags "-X main.gitCommit=$(git rev-list -1 HEAD) -X main.buildEpochSec=$(date +%s)" -o cli/op cli/op.go
RUN cli/op --help
CMD cli/op server
