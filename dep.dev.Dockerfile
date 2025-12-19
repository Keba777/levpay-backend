FROM golang:1.25

ENV GO111MODULE=on
WORKDIR /app
COPY ./go.mod ./go.sum ./

RUN go mod download && go mod tidy && go install github.com/air-verse/air@latest

COPY .air.toml .