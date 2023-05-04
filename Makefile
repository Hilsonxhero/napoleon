## test: runs all tests
test:
	@go test -v ./...

## cover: opens coverage in browser
cover:
	@go test -coverprofile=napoleon.out ./... && go tool cover -html=napoleon.out

## coverage: displays test coverage
coverage:
	@go test -cover ./...

## build_cli: builds the command line tool celeritas and copies it to myapp
build_cli:
	@go build -o ../myapp/celeritas ./cmd/cli