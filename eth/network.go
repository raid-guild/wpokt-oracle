package eth

import (
	eth "github.com/dan13ram/wpokt-oracle/eth/client"
)

func ValidateNetwork() {
	eth.Client.ValidateNetwork()
}
