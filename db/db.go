package db

type DB interface {
	TransactionDB
	MessageDB
	NodeDB
	RefundDB
	SequenceDB
	LockDB
}

type db struct {
	transactionDB
	messageDB
	nodeDB
	refundDB
	sequenceDB
	lockDB
}

func NewDB() DB {
	return &db{}
}
