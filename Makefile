help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

SHELL := /bin/bash
DEPLOY_USER ?= root

build:
	go build -o ./tofui

clean:
	rm -rf ~/.tofui/db

start:
	go run cmd/main.go run -c examples/config.yaml

deploy:
	@echo "Deploying as user: ${DEPLOY_USER}"
	ssh ${DEPLOY_USER}@${DEPLOY_HOST} "sudo systemctl stop tofui"
	scp ./tofui ${DEPLOY_USER}@${DEPLOY_HOST}:/usr/bin/
	scp tofui.yaml ${DEPLOY_USER}@${DEPLOY_HOST}:/etc/tofui/config.yaml
	ssh ${DEPLOY_USER}@${DEPLOY_HOST} "sudo systemctl daemon-reload && sudo systemctl start tofui"

