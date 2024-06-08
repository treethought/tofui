help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

SHELL := /bin/bash
DEPLOY_USER ?= root

build:
	go build -o ./tofui

clean:
	rm -rf /tmp/tofui

start:
	go run cmd/main.go run -c examples/config.yaml

deploy:
	@echo "Deploying as user: ${DEPLOY_USER}"
	scp ./tofui ${DEPLOY_USER}@${DEPLOY_HOST}:/usr/bin/
	scp config.yaml ${DEPLOY_USER}@${DEPLOY_HOST}:/etc/tofui/config.yaml

