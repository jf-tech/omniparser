FROM golang:1.14
WORKDIR /omniparser
COPY . .
# when Dockerfile building in heroku, .git isn't available (removed by heroku slug compiler)
# so git command will fail. in this case, just put "" empty string into the var, we'll figure
# out the commit another way in code.
RUN go build -ldflags "-X main.gitCommit=$(git rev-parse HEAD 2>/dev/null || echo) -X main.buildEpochSec=$(date +%s)" -o cli/op cli/op.go
RUN cli/op --help
CMD cli/op server
