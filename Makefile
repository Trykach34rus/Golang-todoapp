include .env
export

export PROJECT_ROOT=$(shell pwd)

# ------------------------------------------------------------
# env-up: запустить контейнер с БД
# ------------------------------------------------------------
env-up:
	@docker compose up -d todoapp-postgres


# ------------------------------------------------------------
# env-down: остановить контейнер БД
# ------------------------------------------------------------
env-down:
	@docker compose down todoapp-postgres


# ------------------------------------------------------------
# port-forwarder-up: запустить форвардер портов
# ------------------------------------------------------------
env-port-forwarder-up:
	@docker compose up -d port-forwarder


# ------------------------------------------------------------
# port-forwarder-down: остановить форвардер
# ------------------------------------------------------------
env-port-forwarder-down:
	@docker compose down port-forwarder


# ------------------------------------------------------------
# env-cleanup: удалить окружение и данные postgres
# ------------------------------------------------------------
env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down todoapp-postgres port-forwarder && \
		rm -rf $(PROJECT_ROOT)/out/pgdata && \
		echo "Файлы окружения удалены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi


# ------------------------------------------------------------
# migrate-create: создать миграцию
# пример:
# make migrate-create seq=init
# ------------------------------------------------------------
migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm \
		--network todoapp-network \
		todoapp-postgres-migrate \
		-create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"


# ------------------------------------------------------------
# migrate-action: выполнить миграцию
# пример:
# make migrate-action action=up
# ------------------------------------------------------------
migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр для action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm \
		todoapp-postgres-migrate \
		-path=/migrations \
		-database="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@todoapp-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		$(action)


# ------------------------------------------------------------
# применить миграции
# ------------------------------------------------------------
migrate-up:
	@$(MAKE) migrate-action action=up


# ------------------------------------------------------------
# откатить последнюю миграцию
# ------------------------------------------------------------
migrate-down:
	@$(MAKE) migrate-action action=down


# ------------------------------------------------------------
# очистка логов
# ------------------------------------------------------------
log-cleanup:
	@read -p "Очистить все log файлы? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down todoapp-postgres port-forwarder && \
		rm -rf $(PROJECT_ROOT)/out/logs && \
		echo "Логи удалены"; \
	else \
		echo "Очистка отменена"; \
	fi


# ------------------------------------------------------------
# запуск приложения локально
# ------------------------------------------------------------
todoapp-run:
	@export LOGGER_FOLDER=$(PROJECT_ROOT)/out/logs && \
	export POSTGRES_HOST=localhost && \
	go mod tidy && \
	go run $(PROJECT_ROOT)/cmd/todoapp/main.go


# ------------------------------------------------------------
# деплой приложения через docker
# ------------------------------------------------------------
todoapp-deploy:
	@docker compose up -d --build todoapp


# ------------------------------------------------------------
# остановить приложение
# ------------------------------------------------------------
todoapp-undeploy:
	@docker compose down todoapp


# ------------------------------------------------------------
# статус контейнеров
# ------------------------------------------------------------
ps:
	@docker compose ps