.PHONY: migrate migrate_down start

migrate:
	migrate -source file://migrations \
		-database postgres://postgres:secret@localhost:5432/postgres?sslmode=disable up

migrate_down:
	migrate -source file://migrations \
		-database postgres://postgres:secret@localhost:5432/postgres?sslmode=disable down

start:
	reflex -s go run cmd/goreddit/main.go
