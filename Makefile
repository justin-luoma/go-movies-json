default:
	@echo "=============building Local API============="
	docker build -f Dockerfile -t go-movies-json .

up: default
	@echo "=============starting api locally============="
	docker-compose up -d

logs:
	docker-compose logs -f

down:
	docker-compose down

test:
	go test -v -cover ./...

clean: down
	@echo "=============cleaning up============="
	rm -f go-movies-json
	docker system prune -f
	docker volume prune -f
