FROM golang:1.19 AS build

RUN mkdir /home/sparta
WORKDIR /home/sparta

ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build and move binary (we always call sparta the binary and then rename on last step)
RUN CGO_ENABLED=0 go build -o sparta -ldflags="-s -w" main.go
RUN cp sparta /

FROM alpine:3.8

RUN apk update
RUN apk upgrade
RUN apk add ca-certificates && update-ca-certificates
RUN apk add --update tzdata
RUN rm -rf /var/cache/apk/*

COPY docker/config /home/config

COPY --from=build /sparta /home/sparta_ms_--service-name--
RUN ln -s /home/sparta_ms_--service-name-- /usr/bin/sparta_ms_--service-name--

WORKDIR /home
ENTRYPOINT ./sparta_ms_--service-name--
