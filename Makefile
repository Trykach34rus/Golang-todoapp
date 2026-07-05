include .env
export

export PROJECT_ROOT = ${shell pwd}

# ------------------------------------------------------------
# env-up: запустить контейнер с БД
# ------------------------------------------------------------
env-up:
	@docker-compose up -d todoapp-postgres
# Ручной :
# docker-compose --env-file .env up -d todoapp-postgres

# ------------------------------------------------------------
# env-down: остановить и удалить контейнер + сеть (данные сохраняются)
# ------------------------------------------------------------
env-down:
	@docker-compose down todoapp-postgres
# Ручной эквивалент:
# docker-compose down todoapp-postgres

# ------------------------------------------------------------
# env-cleanup: полная очистка (с подтверждением)
# ------------------------------------------------------------
# port-forwarder-up: запустить форвардер портов
# ------------------------------------------------------------
env-port-forwarder-up:
	@docker-compose up -d port-forwarder
# Ручной эквивалент:
# docker-compose up -d port-forwarder

# ------------------------------------------------------------
# port-forwarder-down: остановить и удалить форвардер
# ------------------------------------------------------------
env-port-forwarder-down:
	@docker-compose down port-forwarder
# Ручной эквивалент:
# docker-compose down port-forwarder

env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker-compose down todoapp-postgres port-forwarder && \
		rm -rf ${PROJECT_ROOT}/out/pgdata && \
		echo "Файлы окружения удалены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi
# Ручной эквивалент (без подтверждения, сразу удаляет всё):
# docker-compose down todoapp-postgres && rm -rf out/pgdata

# ------------------------------------------------------------
# migrate-create: создать новый файл миграции
# ------------------------------------------------------------
migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр для seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker-compose run --rm todoapp-postgres-migrate \
		-create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"
# Ручной  (например, для seq=init):
# docker-compose run --rm todoapp-postgres-migrate -create -ext sql -dir /migrations -seq init

# ------------------------------------------------------------
# migrate-action: применить или откатить миграции (универсальная)
# ------------------------------------------------------------
migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр для action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker-compose run --rm todoapp-postgres-migrate \
		-path=/migrations \
		-database="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@todoapp-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		$(action)
# Ручной  для up:
# docker-compose run --rm todoapp-postgres-migrate -path=/migrations -database="postgres://test-user-123:test-posgres-password-456@todoapp-postgres:5432/test_db?sslmode=disable" up
#
# Ручной для down (откат одной миграции):
# docker-compose run --rm todoapp-postgres-migrate -path=/migrations -database="postgres://test-user-123:test-posgres-password-456@todoapp-postgres:5432/test_db?sslmode=disable" down 1

# ------------------------------------------------------------
# migrate-up: применить все новые миграции
# ------------------------------------------------------------
migrate-up:
	$(MAKE) migrate-action action=up
# Ручной :
# docker-compose run --rm todoapp-postgres-migrate -path=/migrations -database="postgres://test-user-123:test-posgres-password-456@todoapp-postgres:5432/test_db?sslmode=disable" up

# ------------------------------------------------------------
# migrate-down: откатить последнюю миграцию
# ------------------------------------------------------------
migrate-down:
	$(MAKE) migrate-action action=down
# Ручной :
# docker-compose run --rm todoapp-postgres-migrate -path=/migrations -database="postgres://test-user-123:test-posgres-password-456@todoapp-postgres:5432/test_db?sslmode=disable" down 1

todoapp-run:
  @export LOGER_FOLDER = ${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST = localhost &&\
	go mod tidy && \
  go run ${PROJECT_ROOT}/cmd/todoapp/main.go


