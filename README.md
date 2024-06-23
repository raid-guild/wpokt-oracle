# wPOKT Oracle

The wPOKT Oracle is a validator node that facilitates the bridging of POKT tokens from the Shannon upgrade of the POKT network to wPOKT on the Ethereum Mainnet and other EVM networks.

## How It Works

The wPOKT Oracle consists of three parallel services for each network that enable the bridging of POKT tokens. Each service operates on an interval specified in the configuration. Here's an overview of their roles:

1. **Message Monitor:**
   Monitors the network for new transactions and creates a message in the database after confirming them.

2. **Message Signer:**
   Validates pending messages and signs them. For the Cosmos network, it broadcasts the messages. For EVM networks, users are responsible for broadcasting the signed messages themselves.

3. **Message Relayer:**
   Monitors the network for confirmed Cosmos transactions and fulfilled message orders on supported EVM networks. It ensures these transactions are confirmed on the respective networks.

Additionally, the Oracle includes a health service that periodically reports the status of the Golang service and sub-services to the database.

Through these services, the wPOKT Oracle bridges POKT tokens to wPOKT, providing a secure and efficient validation process for the entire ecosystem.

## Installation

No specific installation steps are required. Ensure you have Golang installed locally and access to a MongoDB instance, either locally or remotely.

## Usage

To run the wPOKT Oracle, execute the following command:

```bash
go run .
```

### Configuration

The wPOKT Oracle can be configured in the following ways:

1. **Using a YAML File:**
   - A sample configuration file `config.sample.yml` is provided.
   - Specify the config file using the `--yaml` flag:

    ```bash
    go run . --yaml config.yml
    ```

2. **Using an Env File:**
   - A sample environment file `sample.env` is provided.
   - Specify the env file using the `--env` flag:

    ```bash
    go run . --env .env
    ```

3. **Using Environment Variables:**
   - Set the required environment variables directly in your terminal:

    ```bash
    MNEMONIC="your_mnemonic" MONGODB_URI="your_mongodb_uri" ... go run .
    ```

If both a config file and an env file are provided, the config file will be loaded first, followed by the env file. Non-empty values from the env file or provided through environment variables will take precedence over the corresponding values from the config file.

### Makefile

- **Run using:**

    ```bash
    make dev
    ```

- **Build using:**

    ```bash
    make build
    ```

- **Build Docker image using:**

    ```bash
    make docker_build
    ```

## Valid Memo

The validator node requires transactions on the POKT network to include a valid memo in the format of a JSON string. The memo should have the following structure:

```json
{ "address": "0xC9F2D9adfa6C24ce0D5a999F2BA3c6b06E36F75E", "chain_id": "1" }
```

- `address`: The recipient address on the Ethereum network.
- `chain_id`: The chain ID of the Ethereum network (represented as a string).

Transactions with memos not conforming to this format will not be processed by the validator.

## Docker Image

The wPOKT Oracle is also available as a Docker image hosted on [Docker Hub](https://hub.docker.com/r/dan13ram/wpokt-oracle). You can run the validator in a Docker container using the following command:

```bash
docker run -d --env-file .env docker.io/dan13ram/wpokt-oracle:latest
```

Ensure you have set the required environment variables in the `.env` file or directly in the command above.

## Unit Tests

To assess the individual components of the wPOKT Oracle, follow these steps:

1. **Execute Unit Tests:**
    - Open your terminal and run the unit tests using the command:

        ```shell
        make test
        ```

2. **Check Coverage:**
    - To check the test coverage of the codebase, use the following command:

        ```shell
        make test_coverage
        ```

## End-to-End Tests

To test the core functionalities of the wPOKT Oracle in a controlled environment, follow these steps:

1. **Clone the poktroll Repository:**
    - Open your terminal and clone the `poktroll` repository from GitHub:

    ```bash
    git clone https://github.com/pokt-network/poktroll
    cd poktroll
    ```

    - Follow instructions at [poktroll quickstart](https://dev.poktroll.com/develop/developer_guide/quickstart) for more details on running the poktroll network locally.
    - Allow up to 5 minutes for the local network to be fully operational.

2. **Start Local Network:**
    - Move to your project folder and execute the following commands to set up the local network for our oracle and Ethereum networks:

        ```bash
        make clean
        make localnet_up
        ```

    - Please allow up to 5 minutes for the local network to be fully operational.

3. **Run Tests:**
    - Once the local network is ready, initiate the end-to-end tests using the command:

        ```bash
        make e2e_test
        ```

    - Ensure you have Node.js v18 installed to run the tests.

## License

This project is licensed under the MIT License.
