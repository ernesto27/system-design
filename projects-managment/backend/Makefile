.PHONY: refresh-db-test

refresh-db:
	docker rm -f my-postgres-db || true
	docker run --name  my-postgres-db \
		-e POSTGRES_PASSWORD=yourpassword \
		-d \
		-p 5432:5432 \
		postgres:16 
		@echo "Database refreshed"