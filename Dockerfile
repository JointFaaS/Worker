FROM golang:1.13 

WORKDIR /go/src/app

COPY . .

RUN make worker

RUN mkdir /root/.jfWorker

COPY config.yml /root/.jfWorker/

CMD ["/go/src/app/build/worker"]
