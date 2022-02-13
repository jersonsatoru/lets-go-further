APP_ENV=development
APP_PORT=3000
CORS_TRUSTED_ORIGINS="http://localhost:3001 http://localhost:3002"
DSN=postgres://satoru:satoru@localhost:5432/satoru?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_IDLE_TIME=15m
LIMITER_RPS=2
LIMITER_BURST=4
LIMITER_ENABLED=true
MIGRATION_NAME=
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=587
SMTP_USERNAME=1df438ad3e9e01
SMTP_PASSWORD=9aaa7654c1d251
SMTP_SENDER="Greenlight <jersonsatoru@yahoo.com.br>"

run:
	APP_ENV=${APP_ENV} \
	APP_PORT=${APP_PORT} \
	CORS_TRUSTED_ORIGINS=${CORS_TRUSTED_ORIGINS} \
	DSN=${DSN} \
	DB_MAX_IDLE_CONNS=${DB_MAX_IDLE_CONNS} \
	DB_MAX_IDLE_TIME=${DB_MAX_IDLE_TIME} \
	DB_MAX_OPEN_CONNS=${DB_MAX_OPEN_CONNS} \
	LIMITER_BURST=${LIMITER_BURST} \
	LIMITER_ENABLED=${LIMITER_ENABLED} \
	LIMITER_RPS=${LIMITER_RPS} \
	SMTP_HOST=${SMTP_HOST} \
	SMTP_PASSWORD=${SMTP_PASSWORD} \
	SMTP_PORT=${SMTP_PORT} \
	SMTP_SENDER=${SMTP_SENDER} \
	SMTP_USERNAME=${SMTP_USERNAME} \
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
