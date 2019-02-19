FROM golang:1.8

WORKDIR /go/src/github.com/Red350/sync-song-server

COPY . .

RUN go get -d -v .
RUN go install -v .

CMD /go/bin/sync-song-server

EXPOSE 8080

