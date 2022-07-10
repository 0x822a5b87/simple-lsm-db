package simple_lsm_db

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

const TestPath = "/Users/dhy/tmp/kv_store/test/"

func TestKvStore(t *testing.T) {
	kv, err := NewLsmKvStore(TestPath, 10, 10)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		_ = kv.Set(strconv.Itoa(i), strconv.Itoa(i))
	}

	kv.Close()
	kv, err = NewLsmKvStore(TestPath, 10, 10)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		data, found, _ := kv.Get(strconv.Itoa(i))
		if !found {
			t.Error(fmt.Printf("Key [%d] should found [%d] but not found\n", i, i))
			continue
		}
		if data != strconv.Itoa(i) {
			t.Error(fmt.Printf("expected [%d], got [%s]\n", i, data))
		}
	}

	err = os.RemoveAll(TestPath + "*")
	if err != nil {
		t.Error(err)
	}

}
