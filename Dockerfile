FROM golang:1.21

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/app github.com/axeljohnsson/indeed/cmd/indeed

FROM debian:bookworm-slim

COPY --from=0 /usr/local/bin/app /usr/local/bin/app

EXPOSE 8080

CMD ["app"]
