FROM golang:1.22 as base

WORKDIR /app

# copy the source code
COPY . .
RUN go mod download


# build
RUN GOOS=linux go build -o /bin/oracle

# set environment variables

# General configurations
ENV HEALTH_CHECK_INTERVAL_MS ${HEALTH_CHECK_INTERVAL_MS}
ENV HEALTH_CHECK_READ_LAST_HEALTH ${HEALTH_CHECK_READ_LAST_HEALTH}
ENV LOGGER_LEVEL ${LOGGER_LEVEL}
ENV LOGGER_FORMAT ${LOGGER_FORMAT}
ENV MONGODB_URI ${MONGODB_URI}
ENV MONGODB_DATABASE ${MONGODB_DATABASE}
ENV MONGODB_TIMEOUT_MS ${MONGODB_TIMEOUT_MS}

# Mnemonic (Usually a sensitive value, better to set via Docker secrets or external environment)
ENV MNEMONIC ${MNEMONIC}

# cosmos network
ENV COSMOS_NETWORK_START_BLOCK_HEIGHT ${COSMOS_NETWORK_START_BLOCK_HEIGHT}
ENV COSMOS_NETWORK_CONFIRMATIONS ${COSMOS_NETWORK_CONFIRMATIONS}
ENV COSMOS_NETWORK_RPC_URL ${COSMOS_NETWORK_RPC_URL}
ENV COSMOS_NETWORK_GRPC_ENABLED ${COSMOS_NETWORK_GRPC_ENABLED}
ENV COSMOS_NETWORK_GRPC_HOST ${COSMOS_NETWORK_GRPC_HOST}
ENV COSMOS_NETWORK_GRPC_PORT ${COSMOS_NETWORK_GRPC_PORT}
ENV COSMOS_NETWORK_TIMEOUT_MS ${COSMOS_NETWORK_TIMEOUT_MS}
ENV COSMOS_NETWORK_CHAIN_ID ${COSMOS_NETWORK_CHAIN_ID}
ENV COSMOS_NETWORK_CHAIN_NAME ${COSMOS_NETWORK_CHAIN_NAME}
ENV COSMOS_NETWORK_TX_FEE ${COSMOS_NETWORK_TX_FEE}
ENV COSMOS_NETWORK_BECH32_PREFIX ${COSMOS_NETWORK_BECH32_PREFIX}
ENV COSMOS_NETWORK_COIN_DENOM ${COSMOS_NETWORK_COIN_DENOM}
ENV COSMOS_NETWORK_MULTISIG_ADDRESS ${COSMOS_NETWORK_MULTISIG_ADDRESS}
ENV COSMOS_NETWORK_MULTISIG_PUBLIC_KEYS ${COSMOS_NETWORK_MULTISIG_PUBLIC_KEYS}
ENV COSMOS_NETWORK_MULTISIG_THRESHOLD ${COSMOS_NETWORK_MULTISIG_THRESHOLD}
ENV COSMOS_NETWORK_MESSAGE_MONITOR_ENABLED ${COSMOS_NETWORK_MESSAGE_MONITOR_ENABLED}
ENV COSMOS_NETWORK_MESSAGE_MONITOR_INTERVAL_MS ${COSMOS_NETWORK_MESSAGE_MONITOR_INTERVAL_MS}
ENV COSMOS_NETWORK_MESSAGE_SIGNER_ENABLED ${COSMOS_NETWORK_MESSAGE_SIGNER_ENABLED}
ENV COSMOS_NETWORK_MESSAGE_SIGNER_INTERVAL_MS ${COSMOS_NETWORK_MESSAGE_SIGNER_INTERVAL_MS}
ENV COSMOS_NETWORK_MESSAGE_RELAYER_ENABLED ${COSMOS_NETWORK_MESSAGE_RELAYER_ENABLED}
ENV COSMOS_NETWORK_MESSAGE_RELAYER_INTERVAL_MS ${COSMOS_NETWORK_MESSAGE_RELAYER_INTERVAL_MS}

ENV YAML_FILE ${YAML_FILE}

# ethereum networks (later setup via bash file)
ENV NUM_ETHEREUM_NETWORKS ${NUM_ETHEREUM_NETWORKS}

# create app user
RUN adduser --group --system app

RUN chown -R app:app /bin/oracle

RUN chmod +x /bin/oracle

RUN chown -R app:app /app

RUN mv /app/docker/set_eth_envs.sh /app/set_eth_envs.sh

# switch to app user
USER app

# run
ENTRYPOINT ["/bin/sh", "-c", "/app/set_eth_envs.sh && /bin/oracle ${YAML_FILE:+--yaml $YAML_FILE}"]
