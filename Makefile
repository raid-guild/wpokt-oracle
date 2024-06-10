-include .env

all: run

debug :; LOGGER_FORMAT=text LOGGER_LEVEL=debug go run . --yaml defaults/config.one.yml

run-one :; LOGGER_FORMAT=text go run . --yaml defaults/config.one.yml
run-two :; LOGGER_FORMAT=text go run . --yaml defaults/config.two.yml
run-three :; LOGGER_FORMAT=text go run . --yaml defaults/config.three.yml

run :; go run . --yaml config.one.yml

build :; go build -o wpokt-oracle .

send :; poktrolld tx bank send app1 pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar 2000upokt --note "testing memos" --node tcp://127.0.0.1:36657 --yes

docker-build :; docker buildx build . -t dan13ram/wpokt-oracle:v0.0.1 --file ./docker/Dockerfile

docker-run-one :; YAML_FILE=/app/defaults/config.one.yml docker compose -f docker/docker-compose.yml up

hyperlane :; docker compose -f e2e/docker-compose-hyperlane.yml up
