PORT=3000
ENV=development
DSN=postgres://satoru:satoru@localhost:5432/satoru?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_IDLE_TIME=15m
LIMITER_RPS=2
LIMITER_BURST=4
LIMITER_ENABLED=true
MIGRATION_NAME=

run:
	APP_PORT=${PORT} \
	APP_ENV=${ENV} \
	DSN=${DSN} \
	DB_MAX_OPEN_CONNS=${DB_MAX_OPEN_CONNS} \
	DB_MAX_IDLE_CONNS=${DB_MAX_IDLE_CONNS} \
	DB_MAX_IDLE_TIME=${DB_MAX_IDLE_TIME} \
	LIMITER_BURST=${LIMITER_BURST} \
	LIMITER_ENABLED=${LIMITER_ENABLED} \
	LIMITER_RPS=${LIMITER_RPS} \
	go run ./cmd/api/
start_pg:
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
create_migration: 
	migrate create 
		-seq \
		-ext=.sql \
		-dir=./migrations \
		$(MIGRATION_NAME)
run_migration:
	migrate \
	 -path=./migrations \
	 -database="$(DSN)" \
	 up
