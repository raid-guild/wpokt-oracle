-include .env

all: run

debug :; LOGGER_FORMAT=text LOGGER_LEVEL=debug go run . --yaml config.test.yml

run-one :; LOGGER_FORMAT=text go run . --yaml config.one.yml
run-two :; LOGGER_FORMAT=text go run . --yaml config.two.yml
run-three :; LOGGER_FORMAT=text go run . --yaml config.three.yml

run :; go run . --yaml config.test.yml

build :; go build -o wpokt-oracle .

send :; poktrolld tx bank send app1 pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar 2000upokt --note "testing memos" --node tcp://127.0.0.1:36657 --yes
