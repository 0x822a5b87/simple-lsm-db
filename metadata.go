package simple_lsm_db

// FileMetadata represent the lsm structure metadata
type FileMetadata struct {
	version    string
	dataStart  uint
	dataLen    uint
	indexStart uint
	indexLen   uint
}
