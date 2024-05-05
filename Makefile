-include .env

.PHONY: debug

all: debug

debug :; LOG_LEVEL=debug go run . --yaml config.test.yml


build :; go build -o oracle .
