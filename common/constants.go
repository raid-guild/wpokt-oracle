package common

const (
	CollectionLocks        = "locks"
	CollectionTransactions = "transactions"
	CollectionRefunds      = "refunds"
	CollectionMessages     = "messages"
	CollectionNodes        = "nodes"
)

const (
	HyperlaneVersion = 0 // TODO: update to latest when contracts are upgraded
)

const (
	AddressLength          = 20
	CosmosPublicKeyLength  = 33
	DefaultEntropySize     = 256
	DefaultBIP39Passphrase = ""
	DefaultCosmosHDPath    = "m/44'/118'/0'/0/0"
	DefaultETHHDPath       = "m/44'/60'/0'/0/0"
	ZeroAddress            = "0x0000000000000000000000000000000000000000"
)
