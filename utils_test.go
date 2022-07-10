package simple_lsm_db

import (
	"fmt"
	"io"
	"os"
	"testing"
)

const (
	IntTestPath = "/Users/dhy/tmp/kv_store/int"
)

func init() {
	_ = os.Remove(IntTestPath + "*")
}

func TestWriteIntAndReadInt(t *testing.T) {
	file, err := os.OpenFile(IntTestPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Error(err)
	}
	var i int64
	for i = 0; i < 10; i++ {
		err := writeInt64(file, i)
		if err != nil {
			t.Error(err)
		}
	}

	file.Seek(0, io.SeekStart)
	if err != nil {
		t.Error(err)
	}
	for i = 0; i < 10; i++ {
		value, err := readInt64(file)
		if err != nil {
			t.Error(err)
		}
		if value != i {
			t.Error(fmt.Printf("expected [%d], got [%d]\n", i, value))
		}
	}
}

func TestBinary(t *testing.T) {
	var value int64 = 100
	data, err := int64ToBytes(value)
	if err != nil {
		t.Error(err)
	}
	reconvertData, err := bytesToInt64(data)
	if err != nil {
		t.Error(err)
	}

	if value != reconvertData {
		t.Error(fmt.Printf("expected [%d], got [%d]\n", value, reconvertData))
	}
}
