FROM golang:1.13 AS build

WORKDIR /go/src/app

COPY . .

RUN make worker

FROM alpine:3

WORKDIR /root/

RUN mkdir .jfWorker

COPY config.yml .jfWorker/

COPY --from=build /go/src/app/build/ .

CMD ["/root/worker"]