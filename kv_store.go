package simple_lsm_db

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emirpasic/gods/maps/treemap"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
	"sync"
)

var logger *zap.SugaredLogger

const (
	Wal    string = "wal"
	TmpWal string = "wal.tmp"
	Table  string = ".table"
)

type KvStore interface {
	// Get retrieve value with Key
	// return data, success, err
	Get(key string) (string, bool, error)
	// Set value with Key
	Set(key, val string) error
	// Del delete value with Key
	Del(key string) error
	// Close kv store
	Close()
}

type LsmKvStore struct {
	indexLock *sync.RWMutex
	index     *treemap.Map
	tmpIndex  *treemap.Map
	wal       *os.File
	tmpWal    *os.File
	trees     []LsmTree

	path      string
	cacheSize uint
	partSize  uint
}

// NewLsmKvStore new lsm kv store
// path store file path
// cacheSize when in-memory size greater than cacheSize, data will be written to disk
// partSize sparse index region size
func NewLsmKvStore(path string, cacheSize, partSize uint) (kv LsmKvStore, err error) {
	if !strings.HasSuffix(path, "/") {
		return LsmKvStore{}, errors.New("path should end with /")
	}

	kv = LsmKvStore{path: path, cacheSize: cacheSize, partSize: partSize}
	// init lock
	kv.indexLock = &sync.RWMutex{}

	// init index
	kv.index = treemap.NewWithStringComparator()

	// init wal
	kv.wal, err = os.OpenFile(path+Wal, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return kv, errors.New("open wal error : " + err.Error())
	}

	// restore from wal
	err = kv.restoreFromWal(kv.wal)
	if err != nil {
		return kv, err
	}

	// if wal.tmp not exist, we do nothing
	// else we should restore index from tmp wal file,
	// mostly caused by occurring some error during switch index
	if _, err = os.Stat(path + TmpWal); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			// if there is a tmp wal file, mostly caused by occurring some error during switch index
			kv.tmpWal, err = os.OpenFile(path+TmpWal, os.O_RDWR, 0644)
			err = kv.restoreFromWal(kv.tmpWal)
			if err != nil {
				return kv, err
			}
			_ = kv.deleteTmpWal()
		} else {
			return kv, errors.New("open wal.tmp error : " + err.Error())
		}
	}

	// TODO init LsmTree

	return kv, nil
}

func (kv LsmKvStore) Get(key string) (string, bool, error) {
	kv.indexLock.RLock()
	defer kv.indexLock.RUnlock()
	data, found, err := findInTreeMap(kv.index, key)
	if found {
		return data, found, err
	}
	data, found, err = findInTreeMap(kv.tmpIndex, key)
	if found {
		return data, found, err
	}
	// TODO query data from LsmTree

	return "", false, nil
}

func (kv LsmKvStore) Set(key, val string) error {
	kv.indexLock.Lock()
	defer kv.indexLock.Unlock()

	sc := NewSetCommand(key, val)
	return kv.writeData(sc)
}

func (kv LsmKvStore) Del(key string) error {
	//TODO implement me
	panic("implement me")
}

func (kv LsmKvStore) Close() {
	_ = kv.wal.Close()
	_ = kv.deleteTmpWal()
}

func (kv LsmKvStore) writeData(c command) error {
	data, err := json.Marshal(c)
	var j setCommand
	err = json.Unmarshal(data, &j)
	if err != nil {
		logger.Error("marshal value error : ", "error", err.Error())
		return err
	}

	// write wal
	err = writeInt64(kv.wal, int64(len(data)))
	if err != nil {
		return err
	}
	length, err := kv.wal.Write(data)
	if err != nil {
		return err
	}
	if length != len(data) {
		return errors.New("write data error")
	}

	// TODO implement me
	return nil
}

func parseCommandValue(c command) (string, bool) {
	set, ok := c.(setCommand)
	if ok {
		return set.Val, true
	}
	return "", false
}

func findInTreeMap(index *treemap.Map, key string) (string, bool, error) {
	if index == nil {
		return "", false, nil
	}
	c, found := index.Get(key)
	if found {
		// command maybe setCommand or delCommand
		value, b := parseCommandValue(c.(command))
		return value, b, nil
	}
	return "", false, nil
}

func (kv LsmKvStore) deleteTmpWal() error {
	kv.tmpWal = nil
	return os.Remove(kv.path + TmpWal)
}

// restoreFromWal restore index from wal
// this method will update index, so it should Lock() when invoked
func (kv LsmKvStore) restoreFromWal(wal *os.File) error {
	logger.Info("restore from index : ", wal.Name())
	var readError error
	// read util EOF
	for {
		var length int64
		length, readError = readInt64(wal)
		if readError != nil {
			if errors.Is(readError, io.EOF) {
				// this is what we expect, return
				return nil
			} else {
				// that is not what we expect if encounter an error that are not EOF, we should throw it
				return readError
			}
		}
		var data = make([]byte, length)
		_, err := wal.Read(data)
		if err != nil {
			return err
		}
		var jsonMap map[string]interface{}
		err = json.Unmarshal(data, &jsonMap)
		if err != nil {
			return err
		}
		cmd, err := mapToCommand(data, jsonMap)
		switch cmd.commandType() {
		case SET:
			setCmd := cmd.(setCommand)
			kv.index.Put(setCmd.Key, setCmd)
		case DEL:
			delCmd := cmd.(delCommand)
			kv.index.Put(delCmd.Key, delCmd)
		default:
			panic(fmt.Sprintf("invalid command type : [%b]", cmd.commandType()))
		}
	}
}

func init() {
	example, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = example.Sugar()
}
