package common

var CosmosSupportedChainIDs = map[string]bool{
	"poktroll": true,
}

var EthereumSupportedChainIDs = map[string]bool{
	"31337": true,
}

const (
	CollectionLocks        = "locks"
	CollectionTransactions = "transactions"
	CollectionRefunds      = "refunds"
	CollectionMessages     = "messages"
	CollectionNodes        = "nodes"
)
