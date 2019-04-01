FROM golang:1.11

WORKDIR /go/src/github.com/Red350/sync-song-server

COPY . .

# Install the server.
RUN go get -d -v .
RUN go install -v .

# Start the server.
CMD /go/bin/sync-song-server

EXPOSE 8080

