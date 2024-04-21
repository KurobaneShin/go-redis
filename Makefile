run: build
	@./bin/go-redis --listenAddr :5001

build:
	@go build -o bin/go-redis .

