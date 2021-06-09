.PHONY: migrate

migrate:
	migrate -source file://migrations \
		-database postgres://postgres:secret@localhost:5432/postgres?sslmode=disable up

