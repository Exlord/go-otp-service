.PHONY: run build docker-build docker-run

run:
	go run ./cmd/server

build:
	go build -o bin/otpservice ./cmd/server

docker-build:
	docker build -t otpservice:latest .

docker-run:
	docker run --rm -p 8080:8080 -e JWT_SECRET=devsecret otpservice:latest
