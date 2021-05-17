#FROM golang:1.14 as build
#
#WORKDIR /go/src/github.com/genius/
#
#COPY . .
#
#RUN go build -o /go/bin/genius cmd/main.go

FROM ubuntu:18.04

COPY bin/genius /usr/bin/genius

CMD [ "genius" ]