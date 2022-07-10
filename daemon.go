package simple_lsm_db

import "log"

// WalCleaner daemon thread response to clean wal
type WalCleaner struct {
}

func (w WalCleaner) startup() {
	log.Println("wal cleaner startup")
}
