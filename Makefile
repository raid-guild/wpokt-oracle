-include .env

all: run

debug :; LOGGER_LEVEL=debug go run . --yaml config.local.yml

run :; go run . --yaml config.local.yml

build :; go build -o wpokt-oracle .
