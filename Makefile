build:
	@go build -o bin/exchange

run: build
	@bin/exchange

test:
	go test -v ./...

instal-ganache:
	npm install ganache --global

run-ganache:
	ganache -d