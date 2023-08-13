.PHONY: build test vet fmt clean

build:
	@go build ./...

test:
	@go test ./...

vet:
	@go vet ./...

fmt:
	@go fmt ./...

clean:
	@rm -f indeed
