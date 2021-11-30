.PHONY: release

lint:
	golangci-lint run -v

test:
	go test -v ./... -race -cover -covermode=atomic

ci-initialize:
	docker-compose -f ./docker-compose.ci.yaml build

ci-lint:
	docker-compose -f ./docker-compose.ci.yaml run ergo-ci make lint

ci-test:
	docker-compose -f ./docker-compose.ci.yaml run ergo-ci make test

release:
	goreleaser build --rm-dist --snapshot
