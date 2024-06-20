package db

import (
	"fmt"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"

	log "github.com/sirupsen/logrus"
)

// Unlock unlocks a resource
func Unlock(lockID string) error {
	return MongoDB.Unlock(lockID)
}

func LockWriteTransaction(txDoc *models.Transaction) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionTransactions, txDoc.ID.Hex())
	lockID, err = MongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking transaction")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked transaction")
	return
}

func LockWriteRefund(refundDoc *models.Refund) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionRefunds, refundDoc.ID.Hex())
	lockID, err = MongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking refund")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked refund")
	return
}

func LockWriteMessage(messageDoc *models.Message) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionMessages, messageDoc.ID.Hex())
	lockID, err = MongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking message")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked message")
	return
}

const sequenceResourseID = "comsos_sequence"

func LockReadSequences() (lockID string, err error) {
	lockID, err = MongoDB.SLock(sequenceResourseID)
	if err != nil {
		log.WithError(err).Error("Error locking max sequence")
		return
	}
	log.WithField("resource_id", sequenceResourseID).Debug("Locked read sequences")
	return
}

func LockWriteSequence() (lockID string, err error) {
	lockID, err = MongoDB.SLock(sequenceResourseID)
	if err != nil {
		log.WithError(err).Error("Error locking max sequence")
		return
	}
	log.WithField("resource_id", sequenceResourseID).Debug("Locked write sequence")
	return
}
