.PHONY: test

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Get dependencies
	go get -d -v ./...

test: ## Run tests
	go test -short ./...

test-coverage: ## Create a code coverage report
	mkdir -p .cover
	go test -coverprofile .cover/cover.out ./...

test-coverage-html: ## Create a code coverage report in HTML
	mkdir -p .cover
	go test -coverprofile .cover/cover.out ./...
	go tool cover -html .cover/cover.out

lint: gometalinter ## Run linters
	! goimports -d . | grep -vF 'No Exceptions'

fmt: ## Fix formatting issues
	goimports -w .

gometalinter:
	gometalinter --disable-all --enable=golint ./...
