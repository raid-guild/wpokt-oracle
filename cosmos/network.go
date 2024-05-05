package cosmos

import (
	pokt "github.com/dan13ram/wpokt-oracle/cosmos/client"
)

func ValidateNetwork() {
	pokt.Client.ValidateNetwork()
}
