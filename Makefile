.DEFAULT_GOAL := build

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
# Helper
--guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set: $*=abc123 make $(MAKECMDGOALS)"; \
		exit 1; \
	fi

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
	@$(SHELL) -c "cd todo-service-gin; APP_DB_USERNAME=$(PG_USER) \
		APP_DB_PASSWORD=$(PG_PASS) APP_DB_NAME=postgres ./$(BINARY)"

# Tests
test-fake-gin:
	@$(SHELL) -c "cd todo-service-gin; go test -v -tags=fake ./test"

# Dagger
dagger-build:
	@$(SHELL) -c "cd todo-service-gin; BINARY_NAME=$(BINARY) dagger run go run ci/main.go"

dagger-publish:
	@$(SHELL) -c "cd todo-service-gin; DAGGER_PUBLISH=1 DAGGER_REGISTRY_URL=localhost:4567/root/showcase-dagger-golang \
		DAGGER_IMAGE=todo-showcase DAGGER_TAG=0.1 BINARY_NAME=$(BINARY) dagger run go run ci/main.go"

dagger-publish-docker: --guard-REGISTRY_USER --guard-REGISTRY_TOKEN
	@$(SHELL) -c "cd todo-service-gin; DAGGER_PUBLISH=1 DAGGER_REGISTRY_URL=docker.io/$(USER) \
		DAGGER_REGISTRY_USER=$(REGISTRY_USER) DAGGER_REGISTRY_TOKEN=$(REGISTRY_TOKEN) \
		DAGGER_IMAGE=showcase-dagger-golang DAGGER_TAG=0.1 BINARY_NAME=$(BINARY) dagger run go run ci/main.go"

# Helper
clear:
	rm -rf todo-service-gin/$(BINARY)

install:
	go install braces.dev/errtrace/cmd/errtrace@latest
	go install golang.org/x/tools/cmd/deadcode@latest
	go install dagger.io/dagger@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Git
# TOKEN=abc123 make test-build
test-build: --guard-TOKEN
	curl -k -X POST --fail -F token=$(TOKEN) -F ref=master \
		https://localhost:10443/api/v4/projects/2/trigger/pipeline

clone:
	git clone -c http.sslVerify=false \
		https://localhost:10443/root/showcase-dagger-golang.git