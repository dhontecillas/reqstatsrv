build:
	go build -o reqstatsrv ./cmd/reqstatsrv
.PHONY: build

docker:
	docker build -t dhontecillas/reqstatsrv:latest \
		-t dhontecillas/reqstatsrv:v0.2 . 
.PHONY: docker


dockerrun:
	docker run --rm -p 9876:9876 \
		-v ./example:/etc/reqstatsrv \
		dhontecillas/reqstatsrv:v0.2  /reqstatsrv /etc/reqstatsrv/config/example.dockerized.json
.PHONY: dockerrun

test:
	go test ./...
.PHONY: test
