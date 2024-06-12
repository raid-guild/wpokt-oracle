package db

import (
	"fmt"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"

	log "github.com/sirupsen/logrus"
)

// XLock locks a resource for exclusive access
// func XLock(resourceID string) (string, error) {
// 	return mongoDB.XLock(resourceID)
// }
//
// // SLock locks a resource for shared access
// func SLock(resourceID string) (string, error) {
// 	return mongoDB.SLock(resourceID)
// }

// Unlock unlocks a resource
func Unlock(lockID string) error {
	return mongoDB.Unlock(lockID)
}

func LockWriteRefund(refundDoc *models.Refund) (lockID string, err error) {
	resourceID := fmt.Sprintf("%s/%s", common.CollectionRefunds, refundDoc.ID.Hex())
	lockID, err = mongoDB.XLock(resourceID)
	if err != nil {
		log.WithError(err).Error("Error locking refund")
		return
	}
	log.WithField("resource_id", resourceID).Debug("Locked refund")
	return
}

func LockWriteMessage(messageDoc *models.Message) (lockID string, err error) {
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

func LockReadSequences() (lockID string, err error) {
	lockID, err = mongoDB.SLock(sequenceResourseID)
	if err != nil {
		log.WithError(err).Error("Error locking max sequence")
		return
	}
	log.WithField("resource_id", sequenceResourseID).Debug("Locked read sequences")
	return
}

func LockWriteSequence() (lockID string, err error) {
	lockID, err = mongoDB.SLock(sequenceResourseID)
	if err != nil {
		log.WithError(err).Error("Error locking max sequence")
		return
	}
	log.WithField("resource_id", sequenceResourseID).Debug("Locked write sequence")
	return
}
