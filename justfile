go := env_var_or_default('GOCMD', 'go')

default: tidy test

tidy:
	{{go}} mod tidy
	find . -maxdepth 1 -name '*.go' -exec goimports -l -w '{}' ';'
	find . -maxdepth 1 -name '*.go' -exec gofumpt -l -w '{}' ';'
	{{go}} fmt . ./go_test

test:
	{{go}} vet
	golangci-lint run .
	{{go}} test -race -coverprofile=cover.out -timeout=120s
	{{go}} tool cover -html=cover.out -o=cover.html

testloop:
	#!/bin/bash -e
	while :; do
		just test
	done

todo:
	-git grep -e TODO --and --not -e ignoretodo | grep -v ^api/swaggerui | grep -v map:
