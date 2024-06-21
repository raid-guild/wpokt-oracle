package db

import (
	"fmt"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"

	log "github.com/sirupsen/logrus"
)

type LockDB interface {
	Unlock(lockID string) error
	LockWriteTransaction(txDoc *models.Transaction) (lockID string, err error)
	LockWriteRefund(refundDoc *models.Refund) (lockID string, err error)
	LockWriteMessage(messageDoc *models.Message) (lockID string, err error)
	LockReadSequences() (lockID string, err error)
	LockWriteSequence() (lockID string, err error)
}

// Unlock unlocks a resource
func unlock(lockID string) error {
	return mongoDB.Unlock(lockID)
}

func lockWriteTransaction(txDoc *models.Transaction) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionTransactions, txDoc.ID.Hex())
	lockID, err = mongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking transaction")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked transaction")
	return
}

func lockWriteRefund(refundDoc *models.Refund) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionRefunds, refundDoc.ID.Hex())
	lockID, err = mongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking refund")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked refund")
	return
}

func lockWriteMessage(messageDoc *models.Message) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionMessages, messageDoc.ID.Hex())
	lockID, err = mongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking message")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked message")
	return
}

const sequenceResourseID = "comsos_sequence"

func lockReadSequences() (lockID string, err error) {
	lockID, err = mongoDB.SLock(sequenceResourseID)
	if err != nil {
		log.WithError(err).Error("Error locking max sequence")
		return
	}
	log.WithField("resource_id", sequenceResourseID).Debug("Locked read sequences")
	return
}

func lockWriteSequence() (lockID string, err error) {
	lockID, err = mongoDB.SLock(sequenceResourseID)
	if err != nil {
		log.WithError(err).Error("Error locking max sequence")
		return
	}
	log.WithField("resource_id", sequenceResourseID).Debug("Locked write sequence")
	return
}

type lockDB struct{}

func (db *lockDB) Unlock(lockID string) error {
	return unlock(lockID)
}

func (db *lockDB) LockWriteTransaction(txDoc *models.Transaction) (lockID string, err error) {
	return lockWriteTransaction(txDoc)
}

func (db *lockDB) LockWriteRefund(refundDoc *models.Refund) (lockID string, err error) {
	return lockWriteRefund(refundDoc)
}

func (db *lockDB) LockWriteMessage(messageDoc *models.Message) (lockID string, err error) {
	return lockWriteMessage(messageDoc)
}

func (db *lockDB) LockReadSequences() (lockID string, err error) {
	return lockReadSequences()
}

func (db *lockDB) LockWriteSequence() (lockID string, err error) {
	return lockWriteSequence()
}
