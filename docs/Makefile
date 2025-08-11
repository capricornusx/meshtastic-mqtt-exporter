.PHONY: help build serve clean

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать документацию
	docker run --rm -v $(PWD):/docs python:3.11-slim sh -c "cd /docs && pip install -r requirements.txt && mkdocs build --clean"

serve: ## Запустить dev сервер (http://localhost:8000)
	docker run --rm -p 8000:8000 -v $(PWD):/docs python:3.11-slim sh -c "cd /docs && pip install -r requirements.txt && mkdocs serve --dev-addr=0.0.0.0:8000"

clean: ## Очистить собранную документацию
	rm -rf site/

check: ## Проверить конфигурацию
	docker run --rm -v $(PWD):/docs python:3.11-slim sh -c "cd /docs && pip install -r requirements.txt && mkdocs build --verbose --clean"

deploy: ## Деплой на GitHub Pages (только для CI)
	docker run --rm -v $(PWD):/docs python:3.11-slim sh -c "cd /docs && pip install -r requirements.txt && mkdocs gh-deploy --force"