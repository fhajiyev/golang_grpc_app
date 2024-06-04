FROM golang:1.14.2

RUN mkdir -p /go/src/github.com/Buzzvil/buzzscreen-api
RUN mkdir -p /go/bin /go/logs /go/shared

RUN go get github.com/derekparker/delve/cmd/dlv

WORKDIR /go

EXPOSE 8081
EXPOSE 2345
