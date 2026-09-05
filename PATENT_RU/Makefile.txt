.PHONY: fmt test vet build smoke

fmt:
	gofmt -w cmd internal

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/trest ./cmd/trest

smoke: build
	./bin/trest --help
	./bin/trest generate --config proektirovka-sdaniy/configs/court_tosno.yaml --out ./output
