package database

import (
	"os"
	"sync"
)

type journal interface {
	load(s memStorage) error
	checkpoint(s memStorage)
	addRemoveOp(key string)
	addWriteOp(key string, value []byte)
}

type fileJournal struct {
	sync.Mutex

	file *os.File
}

func newJournal() *fileJournal {
	return &fileJournal{}
}

func (fj *fileJournal) load(s memStorage) error {
	return nil
}

func (fj *fileJournal) checkpoint(s memStorage) {
}

func (fj *fileJournal) addRemoveOp(key string) {
}

func (fj *fileJournal) addWriteOp(key string, value []byte) {
}
