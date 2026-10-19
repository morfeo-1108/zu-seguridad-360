SHELL := /bin/bash
.SHELLFLAGS := -euo pipefail -c

GO_SERVICES := core mail notifications autopilot attention onboarding billing sla drone media
.PHONY: help dev setup setup-core-test sync-go-deps sync-go-deps-clean build test test-go test-ai test-company-isolation lint migrate seed clean dev-infra dev-down reset-dev-volumes

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-24s %s\n", $$1, $$2}'

setup: ## Prepara el stack local de pruebas con datos sintéticos
	./scripts/setup.sh

setup-core-test: ## Inicia sólo Postgres + Core y crea datos sintéticos para QA básica
	./scripts/setup-core-test.sh

sync-go-deps: ## Ejecuta tidy, verify y build en todos los módulos Go
	./scripts/sync-go-deps.sh

sync-go-deps-clean: ## Limpia la caché global Go y regenera/verifica/compila todos los módulos
	./scripts/sync-go-deps.sh --clean-cache

dev: ## Levanta backend e infraestructura sólo en Docker local
	./scripts/assert-local-docker.sh
	docker compose -f docker-compose.dev.yml up --build

dev-infra: ## Levanta sólo dependencias de desarrollo
	./scripts/assert-local-docker.sh
	docker compose -f docker-compose.dev.yml up -d postgres redis rabbitmq weaviate neo4j emqx minio

dev-down: ## Detiene el stack local sin borrar datos
	./scripts/assert-local-docker.sh
	docker compose -f docker-compose.dev.yml down

build: ## Compila los módulos Go y verifica sintaxis Python
	@for svc in $(GO_SERVICES); do echo "Building $$svc..."; (cd services/$$svc && go build ./...); done
	python3 -m compileall -q services/ai/app services/documents/app

test: ## Ejecuta todas las suites Go y Python disponibles; cualquier fallo falla el target
	./scripts/run-tests.sh

test-go: ## Ejecuta todas las suites Go
	@for svc in $(GO_SERVICES); do echo "Testing $$svc..."; (cd services/$$svc && go test ./... -count=1); done

test-ai: ## Ejecuta tests Python de IA
	cd services/ai && python3 -m pytest tests/ -v --asyncio-mode=auto

test-company-isolation: ## Prueba A→B con las cuentas sintéticas; requiere Core local activo
	python3 scripts/smoke-company-isolation.py

lint: ## Lint Go (requiere golangci-lint)
	@for svc in $(GO_SERVICES); do echo "Linting $$svc..."; (cd services/$$svc && golangci-lint run); done

migrate: ## Aplica sólo SQL base al Postgres local de Compose; el Core hace AutoMigrate sólo en dev/test
	./scripts/migrate.sh

seed: ## Inserta dos profesionales y empresas sintéticas en Postgres local tras confirmación
	./scripts/seed.sh

clean: ## Limpia artefactos locales; no borra volúmenes de bases de datos
	find . -name '*.out' -delete
	find . -name '__pycache__' -type d -prune -exec rm -rf {} +

reset-dev-volumes: ## BORRA los datos del Compose local sólo con confirmación explícita
	@test "$${CONFIRM_DEV_VOLUME_RESET:-}" = YES || (echo 'Refused: set CONFIRM_DEV_VOLUME_RESET=YES to delete local volumes' >&2; exit 1)
	./scripts/assert-local-docker.sh
	docker compose -f docker-compose.dev.yml down -v
