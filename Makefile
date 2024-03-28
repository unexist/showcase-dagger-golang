.DEFAULT_GOAL := build
.ONESHELL:

BINARY := todo-service.bin
PODNAME := showcase
PG_USER := postgres
PG_PASS := postgres

define JSON_TODO
curl -X 'POST' \
  'http://localhost:8080/todo' \
  -H 'accept: */*' \
  -H 'Content-Type: application/json' \
  -d '{
  "description": "string",
  "done": true,
  "title": "string"
}'
endef
export JSON_TODO

# Tools
todo:
	@echo $$JSON_TODO | bash

list:
	@curl -X 'GET' 'http://localhost:8080/todo' -H 'accept: */*' | jq .

psql:
	@PGPASSWORD=$(PG_PASS) psql -h 127.0.0.1 -U $(PG_USER)

psql-schema:
	@PGPASSWORD=$(PG_PASS) psql -h 127.0.0.1 -U $(PG_USER) -f ./schema.sql

swagger:
	@$(SHELL) -c "cd todo-service-gin; swag init"

open-swagger:
	open http://localhost:8080/swagger/index.html

# Build
build-gin:
	@$(SHELL) -c "cd todo-service-gin; GO111MODULE=on; go mod download; go build -o $(BINARY)"

# Analysis
vet-gin:
	@$(SHELL) -c "cd todo-service-gin; go vet"

# Run
run-gin:
	@$(SHELL) -c "cd todo-service-gin; APP_DB_USERNAME=$(PG_USER) APP_DB_PASSWORD=$(PG_PASS) APP_DB_NAME=postgres ./$(BINARY)"

# Tests
test-fake-gin:
	@$(SHELL) -c "cd todo-service-gin; go test -v -tags=fake ./test"

# Dagger
dagger:
	@$(SHELL) -c "cd todo-service-gin; dagger run go run ci/main.go"

# Helper
clear:
	rm -rf todo-service-gin/$(BINARY)

install:
	go install braces.dev/errtrace/cmd/errtrace@latest
	go install golang.org/x/tools/cmd/deadcode@latest
	go install dagger.io/dagger@latest

# Git
# TOKEN=abc123 make test-bvuild
test-build:
	curl -k -X POST --fail -F token=$(TOKEN) -F ref=master \
		https://localhost:10443/api/v4/projects/2/trigger/pipeline

clone:
	git clone -c http.sslVerify=false \
		https://localhost:10443/root/showcase-dagger-golang.git