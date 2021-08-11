FROM golang:1.16
WORKDIR /go/src/github.com/weaveworks/weave-gitops
COPY .git .git
COPY api api
COPY cmd cmd
COPY manifests manifests
COPY pkg pkg
COPY test test
COPY tools tools
COPY ui ui
COPY go.mod go.mod
COPY go.sum go.sum
COPY package.json package.json
COPY package-lock.json package-lock.json
COPY Makefile Makefile
RUN go get -d -v ./...
ENV SKIP_FETCH_TOOLS=1
CMD ["make","unit-tests"]
