FROM golang:1.16

RUN curl -sL https://deb.nodesource.com/setup_14.x | bash -
RUN apt-get install -y nodejs
WORKDIR /go/src/github.com/weaveworks/weave-gitops
COPY . .
RUN go get -d -v ./...
RUN make all BINARY_NAME=wego
CMD ["./bin/wego"] 
