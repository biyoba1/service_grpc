include .env
LOCAL_BIN:=$(CURDIR)/bin

LOCAL_MIGRATION_DIR=$(MIGRATION_DIR)
LOCAL_MIGRATION_DSN="host=localhost port=$(PG_PORT) dbname=$(PG_DATABASE_NAME) user=$(PG_USER) password=$(PG_PASSWORD) sslmode=disable"

install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.14.0

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate:
	make generate-note-api

generate-note-api:
	mkdir -p pkg/chat_v1
	protoc --proto_path=protobuf/src --proto_path=api/chat_v1 \
    --go_out=pkg/chat_v1 --go_opt=paths=source_relative \
    --plugin=protoc-gen-go=bin/protoc-gen-go \
    --go-grpc_out=pkg/chat_v1 --go-grpc_opt=paths=source_relative \
    --plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
    api/chat_v1/chat.proto

build:
	GOOS=linux GOARCH=amd64 go build -o service_linux cmd/main.go

copy-to-server:
	scp service_linux root@5.44.45.15:

docker-build-and-push:
	docker buildx build --no-cache --platform linux/amd64 --network=host -t cr.selcloud.ru/valera/chat_server:v0.0.1 .
	docker login -u token -p CRgAAAAACouDtA82R6Fjq0mh1qBBgvkDxh46y5PJ cr.selcloud.ru/valera
	docker push cr.selcloud.ru/valera/chat_server:v0.0.1

local-migration-status:
	goose -dir ${LOCAL_MIGRATION_DIR} postgres ${LOCAL_MIGRATION_DSN} status -v

local-migration-up:
	goose -dir ${LOCAL_MIGRATION_DIR} postgres ${LOCAL_MIGRATION_DSN} up -v

local-migration-down:
	goose -dir ${LOCAL_MIGRATION_DIR} postgres ${LOCAL_MIGRATION_DSN} down -v

