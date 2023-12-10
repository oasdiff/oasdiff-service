GOLANGCILINT_VERSION=v1.52.2

.PHONY: devtools
devtools:  ## Install dev tools
	@echo "==> Installing dev tools..."
	brew install golangci-lint

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: docs
docs: ## Generate docs
	docker run -v $(PWD):/spec redocly/cli build-docs docs/openapi.yaml
	mv redoc-static.html docs.html
