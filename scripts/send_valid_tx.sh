#!/bin/bash

chain_id="38545"
address="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

memo="{\"chain_id\":\"$chain_id\",\"address\":\"$address\"}"

echo "Sending 1000 upokt"
echo "Memo: $memo"

poktrolld tx bank send app1 pokt13tsl3aglfyzf02n7x28x2ajzw94muu6y57k2ar 1000upokt --note "$memo" --node tcp://127.0.0.1:36657 --home=/home/dan13ram/code/raid-guild/pocket/poktroll/localnet/poktrolld --yes
