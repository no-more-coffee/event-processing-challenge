.PHONY: all up migrate generate

all: up migrate

up:
	docker compose up -d

migrate:
	docker compose exec database sh -c 'psql -U casino < /db/migrations/00001.create_base.sql'

generator:
	docker compose run --rm generator
