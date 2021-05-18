# Due to networking issues, the commented lines below won't work correctly.

#FROM golang:1.14 as build
#
#WORKDIR /go/src/github.com/genius/
#
#COPY . .
#
#RUN go build -o /go/bin/genius cmd/main.go

FROM golang:1.14

WORKDIR /go/src/genius

# for debugging purpose
COPY . .

COPY bin/genius /usr/bin/genius

CMD [ "genius" ]