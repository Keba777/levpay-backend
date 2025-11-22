HOSTNAME=$(shell cat /etc/hostname)
BUILD_DEV=go build -buildvcs=false -o ./tmp/$(HOSTNAME)
BINARY_PATH=tmp/$(HOSTNAME)
APP_MAIN=./cmd/app/main.go
AUTH_MAIN=./cmd/auth/main.go
USER_MAIN=./cmd/user/main.go
WALLET_MAIN=./cmd/wallet/main.go
TRANSACTION_MAIN=./cmd/transaction/main.go
KYC_MAIN=./cmd/kyc/main.go
FILE_MAIN=./cmd/file/main.go
NOTIFICATION_MAIN=./cmd/notification/main.go
BILLING_MAIN=./cmd/billing/main.go
ADMIN_MAIN=./cmd/admin/main.go
CRON_MAIN=./cmd/cron/main.go
AIR_COMMAND=air -c .air.toml -build.bin $(BINARY_PATH)
BUILD_COMMAND=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./binary
DOCS_OUTPUT_PATH=docs/
DOCS_OUTPUT_TYPES=go,json
DEPENDENCIES_DOCKERFILE_PROD=dep.prod.Dockerfile
DEPENDENCIES_DOCKERFILE_DEV=dep.dev.Dockerfile
DOCKER_COMPOSE_FILE_DEV=docker-compose.dev.yaml
DOCKER_COMPOSE_FILE_PROD=docker-compose.yaml
ENV_FILE_PATH=.env.example
.PHONY: all
##### DEVELOPMENT COMMANDS #####
swagger:
	@swag init --dir cmd/app/,internal/ --outputTypes $(DOCS_OUTPUT_TYPES) --output $(DOCS_OUTPUT_PATH)
fmt:
	@gofmt -l -w .
swag_fmt:
	@swag fmt --dir cmd/app/,internal/
pretty: fmt swag_fmt swagger
##### SERVICES COMMANDS #####
app_prod:
	$(BUILD_COMMAND) $(APP_MAIN)
app_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(APP_MAIN)"
auth_prod:
	$(BUILD_COMMAND) $(AUTH_MAIN)
auth_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(AUTH_MAIN)"
user_prod:
	$(BUILD_COMMAND) $(USER_MAIN)
user_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(USER_MAIN)"
wallet_prod:
	$(BUILD_COMMAND) $(WALLET_MAIN)
wallet_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(WALLET_MAIN)"
transaction_prod:
	$(BUILD_COMMAND) $(TRANSACTION_MAIN)
transaction_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(TRANSACTION_MAIN)"
kyc_prod:
	$(BUILD_COMMAND) $(KYC_MAIN)
kyc_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(KYC_MAIN)"
file_prod:
	$(BUILD_COMMAND) $(FILE_MAIN)
file_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(FILE_MAIN)"
notification_prod:
	$(BUILD_COMMAND) $(NOTIFICATION_MAIN)
notification_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(NOTIFICATION_MAIN)"
billing_prod:
	$(BUILD_COMMAND) $(BILLING_MAIN)
billing_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(BILLING_MAIN)"
admin_prod:
	$(BUILD_COMMAND) $(ADMIN_MAIN)
admin_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(ADMIN_MAIN)"
cron_prod:
	$(BUILD_COMMAND) $(CRON_MAIN)
cron_dev:
	@$(AIR_COMMAND) -build.cmd "$(BUILD_DEV) $(CRON_MAIN)"
##### BUILD COMMANDS #####
build-dependencies-dev:
	docker build -t dependencies_dev -f $(DEPENDENCIES_DOCKERFILE_DEV) .
build_dev:
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) build
build-dependencies-prod:
	docker build -t dependencies_prod -f $(DEPENDENCIES_DOCKERFILE_PROD) .
build_prod: build-dependencies-prod
	docker compose -f $(DOCKER_COMPOSE_FILE_PROD) --env-file $(ENV_FILE_PATH) build
##### COMPOSE COMMANDS #####
dev_up_d: build-dependencies-dev build_dev
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --profile full --env-file $(ENV_FILE_PATH) up -d
dev_up: build-dependencies-dev build_dev
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --profile full --env-file $(ENV_FILE_PATH) up
dev_minimal: build-dependencies-dev
	@echo "Starting minimal development environment (core services only)..."
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) --profile minimal up -d
dev_backend: build-dependencies-dev
	@echo "Starting backend development environment..."
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) --profile backend up -d
dev_app_only: build-dependencies-dev
	@echo "Starting only app service with dependencies..."
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) --profile app --profile backend up -d
dev_rebuild_service:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make dev_rebuild_service SERVICE=<service_name>"; exit 1; fi
	@echo "Rebuilding $(SERVICE) service..."
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) build $(SERVICE)
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) up -d $(SERVICE)
dev_logs:
	@if [ -z "$(SERVICE)" ]; then \
		docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) logs -f; \
	else \
		docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) logs -f $(SERVICE); \
	fi
dev_shell:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make dev_shell SERVICE=<service_name>"; exit 1; fi
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) exec $(SERVICE) sh
dev_test_service:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make dev_test_service SERVICE=<service_name>"; exit 1; fi
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) exec $(SERVICE) go test ./...
dev_down:
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) down
dev_down_v:
	docker compose -f $(DOCKER_COMPOSE_FILE_DEV) --env-file $(ENV_FILE_PATH) down -v
dev_clean: dev_down_v
	@echo "Cleaning up development environment..."
	docker system prune -f
	docker volume prune -f
up: build_prod
	docker compose -f $(DOCKER_COMPOSE_FILE_PROD) --env-file $(ENV_FILE_PATH) up
up_d: build_prod
	docker compose -f $(DOCKER_COMPOSE_FILE_PROD) --env-file $(ENV_FILE_PATH) up -d
down:
	docker compose -f $(DOCKER_COMPOSE_FILE_PROD) --env-file $(ENV_FILE_PATH) down
down_v:
	docker compose -f $(DOCKER_COMPOSE_FILE_PROD) --env-file $(ENV_FILE_PATH) down -v