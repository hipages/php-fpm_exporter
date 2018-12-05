.PHONY: test

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	go test -short ./...

test-coverage:
	mkdir -p .cover
	go test -coverprofile .cover/cover.out ./...

test-coverage-html:
	mkdir -p .cover
	go test -coverprofile .cover/cover.out ./...
	go tool cover -html .cover/cover.out

lint: gometalinter ## Run linters
	! goimports -d . | grep -vF 'No Exceptions'

fmt: ## Fix formatting issues
	goimports -w .

gometalinter:
	CGO_ENABLED=0 gometalinter --disable-all --enable=megacheck --enable=golint --enable=unconvert --enable=vet --enable=vetshadow --vendor ./...
