package pokt

import (
	pokt "github.com/dan13ram/wpokt-oracle/pokt/client"
)

func ValidateNetwork() {
	pokt.Client.ValidateNetwork()
}
