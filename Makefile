-include .env

all: run

debug :; LOGGER_LEVEL=debug go run . --yaml config.test.yml

run :; go run . --yaml config.test.yml

build :; go build -o wpokt-oracle .
