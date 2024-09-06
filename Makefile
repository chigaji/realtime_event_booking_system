# Makefile

# Load environment variables from .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

SQL_FILE=schema.sql  # Name of the SQL file to run
# SQL_FILE=001_create_users_table.sql  # Name of the SQL file to run
DB_CONTAINER=postgres_db  # Name of the PostgreSQL container (must match the name in docker-compose.yml)
REDIS_CONTAINER=redis_db  # Name of the Redis container


run:
	go run ./cmd/server/main.go
# Start the services
up:
	docker-compose up -d

# Stop the services
down:
	docker-compose down

# View logs from the services
logs:
	docker-compose logs -f

# Rebuild and start the services
rebuild:
	docker-compose up -d --build

# Remove containers, networks, volumes, and images created by up
clean:
	docker-compose down -v --rmi all

# Show status of the services
status:
	docker-compose ps

# Run the SQL file against the PostgreSQL database
migrate:
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $(SQL_FILE)

# Exec into the PostgreSQL container
dbshell:
	docker exec -it $(DB_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
