FROM golang:1.18.0

VOLUME /bot
WORKDIR /app

RUN export GO111MODULE=on

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot ./cmd/bot/


CMD ["./bot"]
