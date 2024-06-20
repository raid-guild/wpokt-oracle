package db

import (
	"testing"

	mocks "github.com/dan13ram/wpokt-oracle/db/mocks"
)

var oldDB Database

func MockDatabase(t *testing.T) *mocks.MockDatabase {
	mockDB := mocks.NewMockDatabase(t)
	oldDB = mongoDB
	mongoDB = mockDB

	return mockDB
}

func UnmockDatabase() {
	mongoDB = oldDB
}
