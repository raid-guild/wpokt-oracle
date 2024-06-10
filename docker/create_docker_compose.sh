#!/bin/bash

# Load environment variables from .env file
# set -o allexport
# source .env
# set -o allexport

# Get the number of Ethereum networksm, default to 2
NUM_ETHEREUM_NETWORKS=${NUM_ETHEREUM_NETWORKS:-2}

# Generate dynamic environment variable lines for Ethereum networks
env_vars=""
for i in $(seq 0 $(($NUM_ETHEREUM_NETWORKS - 1))); do
  env_vars+="      ETHEREUM_NETWORKS_${i}_START_BLOCK_HEIGHT: \${ETHEREUM_NETWORKS_${i}_START_BLOCK_HEIGHT}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_CONFIRMATIONS: \${ETHEREUM_NETWORKS_${i}_CONFIRMATIONS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_RPC_URL: \${ETHEREUM_NETWORKS_${i}_RPC_URL}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_TIMEOUT_MS: \${ETHEREUM_NETWORKS_${i}_TIMEOUT_MS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_CHAIN_ID: \${ETHEREUM_NETWORKS_${i}_CHAIN_ID}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_CHAIN_NAME: \${ETHEREUM_NETWORKS_${i}_CHAIN_NAME}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MAILBOX_ADDRESS: \${ETHEREUM_NETWORKS_${i}_MAILBOX_ADDRESS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MINT_CONTROLLER_ADDRESS: \${ETHEREUM_NETWORKS_${i}_MINT_CONTROLLER_ADDRESS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_OMNI_TOKEN_ADDRESS: \${ETHEREUM_NETWORKS_${i}_OMNI_TOKEN_ADDRESS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_WARP_ISM_ADDRESS: \${ETHEREUM_NETWORKS_${i}_WARP_ISM_ADDRESS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_ORACLE_ADDRESSES: \${ETHEREUM_NETWORKS_${i}_ORACLE_ADDRESSES}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MESSAGE_MONITOR_ENABLED: \${ETHEREUM_NETWORKS_${i}_MESSAGE_MONITOR_ENABLED}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MESSAGE_MONITOR_INTERVAL_MS: \${ETHEREUM_NETWORKS_${i}_MESSAGE_MONITOR_INTERVAL_MS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MESSAGE_SIGNER_ENABLED: \${ETHEREUM_NETWORKS_${i}_MESSAGE_SIGNER_ENABLED}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MESSAGE_SIGNER_INTERVAL_MS: \${ETHEREUM_NETWORKS_${i}_MESSAGE_SIGNER_INTERVAL_MS}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MESSAGE_RELAYER_ENABLED: \${ETHEREUM_NETWORKS_${i}_MESSAGE_RELAYER_ENABLED}\n"
  env_vars+="      ETHEREUM_NETWORKS_${i}_MESSAGE_RELAYER_INTERVAL_MS: \${ETHEREUM_NETWORKS_${i}_MESSAGE_RELAYER_INTERVAL_MS}\n"
done

# Substitute the placeholder with actual environment variable lines
sed "s/{{ETHEREUM_NETWORKS_ENV_VARS}}/$env_vars/g" docker-compose.template.yml > docker-compose.yml

# echo "Generated docker-compose.yml:"
cat docker-compose.yml
