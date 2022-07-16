FROM golang:1.16

VOLUME ["/app"]
WORKDIR /app

RUN cd /app && go mod download

EXPOSE 8080

CMD air
