PORT=3000
ENV=development

run:
	go run ./cmd/api/ -port ${PORT} -env ${ENV}