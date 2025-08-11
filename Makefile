init:
	cp .env.example .env
	cp example.config.yml config.yml
	docker-compose up -d
	docker-compose exec go-api bash -c "cd /app && goose up"

rebuild:
	docker-compose up -d --build

up:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down