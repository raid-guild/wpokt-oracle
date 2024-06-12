-include .env

all: clean install test build

.PHONY: dev
dev : dev_one

.PHONY: dev_one
dev_one:; CGO_ENABLED=0 go run . --yaml ./defaults/config.local.one.yml

.PHONY: dev_two
dev_two:; go run . --yaml ./defaults/config.local.two.yml

.PHONY: dev_three
dev_three:; go run . --yaml ./defaults/config.local.three.yml

.PHONY: clean
clean: clean_tmp_data
	go clean && go mod tidy

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
docker_dev : docker_one

.PHONY: docker_one
docker_one :; YAML_FILE=/app/defaults/config.local.one.yml docker compose -f docker/docker-compose.yml up --force-recreate

.PHONY: docker_two
docker_two :; YAML_FILE=/app/defaults/config.local.two.yml docker compose -f docker/docker-compose.yml up --force-recreate

.PHONY: docker_three
docker_three :; YAML_FILE=/app/defaults/config.local.three.yml docker compose -f docker/docker-compose.yml up --force-recreate

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
