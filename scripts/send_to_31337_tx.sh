#!/bin/bash

chain_id="31337"
address="0x0E90A32Df6f6143F1A91c25d9552dCbc789C34Eb"

memo="{\"chain_id\":\"$chain_id\",\"address\":\"$address\"}"

echo "Sending 1000 upokt to 31337"
echo "Memo: $memo"

poktrolld tx bank send app1 pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar 1000upokt --note "$memo" --node tcp://127.0.0.1:36657 --home=/home/dan13ram/code/raid-guild/pocket/poktroll/localnet/poktrolld --yes
