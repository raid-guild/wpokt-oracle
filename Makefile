-include .env

all: clean install test build

.PHONY: debug
dev :; go run . --yaml ./defaults/config.local.one.yml

.PHONY: clean
clean :; go clean && go mod tidy && make clean_tmp_data

.PHONY: clean_tmp_data
clean_tmp_data :; if [ -d "/tmp/data" ]; then sudo rm -rf /tmp/data; fi

.PHONY: install
install :; go mod download && go mod verify

.PHONY: test
test :; go test -v ./...

.PHONY: build
build :; go build -o wpokt-oracle .

.PHONY: docker_build
docker_build :; docker buildx build . -t dan13ram/wpokt-oracle:v0.0.1 --file ./docker/Dockerfile

.PHONY: docker_dev
docker_dev :; YAML_FILE=/app/defaults/config.local.one.yml docker compose -f docker/docker-compose.yml up

.PHONY: localnet_up
localnet_up:; docker compose -f e2e/docker-compose.yml up --force-recreate

.PHONY: prompt_user
prompt_user:
	@echo "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

.PHONY: docker_wipe
docker_wipe: prompt_user ## [WARNING] Remove all the docker containers, images and volumes.
	docker ps -a -q | xargs -r -I {} docker stop {}
	docker ps -a -q | xargs -r -I {} docker rm {}
	docker images -q | xargs -r -I {} docker rmi {}
	docker volume ls -q | xargs -r -I {} docker volume rm {}

.PHONY: e2e_tests
e2e_tests :; cd e2e && yarn install && yarn test
