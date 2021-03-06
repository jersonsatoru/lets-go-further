include .envrc

current_time := $(shell date --iso-8601=seconds)
apiVersion := $(shell git describe --always --dirty --tags --long)

## help: print hist help message
.PHONY: help
help:
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## api/version: display API version
api/version:
	./bin/api -version

## api/build: build api cmd binaries
api/build:
	@echo 'Building ./cmd/api'
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -X main.buildTime=${current_time} -X main.version=${apiVersion}' -o=./bin/api ./cmd/api

## api/run: run the cmd/api application
.PHONY: app/run
api/run:
	go run ./cmd/api/
## db/start: start pg docker image
.PHONY: db/start
db/start:
	docker stop postgres || true
	docker rm postgres || true
	docker run -d \
		--name postgres \
		-p 5432:5432 \
		-v pg-data:/var/lib/postgresql/data \
		-v "${PWD}/init.sql":/docker-entrypoint-initdb.d/init.sql \
		-e POSTGRES_PASSWORD=satoru \
		-e POSTGRES_USER=satoru \
		-e POSTGRES_DATABASE=satoru \
		postgres:14.1-alpine3.15
## db/migration/create name=$1: create a new database migration
.PHONY: db/migration/create
db/migration/create: confirm
	echo ${name}
	migrate create -seq -ext=.sql -dir=./migrations ${name}
## db/migration/up: apply all up database migrations
.PHONY: db/migration/run
db/migration/run:
	migrate -path=./migrations -database="$(DSN)" up
## qc/audit: Quality control, execute code formtat, vetting and static check
.PHONY: qc/audit
qc/audit: qc/vendor
	@echo 'Formatting code'
	go fmt ./...
	@echo 'Vetting code'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## qc/vendor: vendoring third-party packages (storing locally) and normalizing packages
.PHONY: qc/vendor
qc/vendor:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Vendoring third-party packages'
	go mod vendor
