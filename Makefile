VERSION ?= "v0.4"

build:
	go build -o reqstatsrv ./cmd/reqstatsrv
.PHONY: build

docker:
	docker build \
		-t dhontecillas/reqstatsrv:latest \
		-t dhontecillas/reqstatsrv:${VERSION} . 
.PHONY: docker


dockerrun: docker
	docker run --rm -p 9876:9876 \
		-v ./example:/etc/reqstatsrv \
		dhontecillas/reqstatsrv:${VERSION} \
		/reqstatsrv /etc/reqstatsrv/config/example.dockerized.json
.PHONY: dockerrun

test:
	go test ./...
.PHONY: test
