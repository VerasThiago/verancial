all:
	docker-compose up

all_build:
	docker-compose up --build

start_api_local:
	cd api && VERANCIAL_DEPLOY_ENV=local go run main.go

start_report_process_worker_local:
	cd data-process-worker && VERANCIAL_DEPLOY_ENV=local go run main.go

start_app_integration_worker_local:
	cd app-integration-worker && VERANCIAL_DEPLOY_ENV=local go run main.go

start_api_docker:
	docker-compose up api

start_login_local:
	cd login && VERANCIAL_DEPLOY_ENV=local go run main.go

start_login_docker:
	docker-compose up login

start_frontend_local:
	cd frontend && npm run dev

start_frontend_docker:
	docker-compose up frontend

start_redis:
	docker-compose up -d worker-redis

# Local database commands removed - now using Supabase
# To start local database for development, uncomment the database service in docker-compose.yaml and use:
# start_db:
# 	docker-compose up -d database
# ssh_db:
# 	docker exec -it database bash

migrate_db:
	cd shared/scripts && \
	VERANCIAL_DEPLOY_ENV=local go run mage.go migrateUserModel && \
	VERANCIAL_DEPLOY_ENV=local go run mage.go migrateTransactionModel

start_all:
	cd shared/scripts && \
	osascript start_all.sh

