postgres:
	docker run --name awaydb -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it awaydb createdb --username=root --owner=root away

dropdb:
	docker exec -it awaydb dropdb away

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/away?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/away?sslmode=disable" -verbose down

.PHONY: postgres createdb dropdb migrateup migratedown