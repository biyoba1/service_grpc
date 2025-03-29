FROM golang:1.23-alpine AS builder

COPY . /auth/user_api/source
WORKDIR /auth/user_api/source

RUN go mod download
RUN go build -o ./bin/user_server cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /auth/user_api/source/bin/user_server .

CMD ["./user_server"]