FROM golang:1.23-alpine AS builder

COPY . /chat_server/chat_api/source
WORKDIR /chat_server/chat_api/source

RUN go mod download
RUN go build -o ./bin/chat_server cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /chat_server/chat_api/source/bin/chat_server .

CMD ["./chat_server"]