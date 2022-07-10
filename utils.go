package simple_lsm_db

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func writeInt64(file *os.File, value int64) error {
	data, err := int64ToBytes(value)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}

func readInt64(file *os.File) (int64, error) {
	var data = make([]byte, 8)
	_, err := file.Read(data)
	if err != nil {
		return 0, err
	}
	return bytesToInt64(data)
}

//int64ToBytes 整形转换成字节
func int64ToBytes(n int64) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes(), err
}

//bytesToInt64 字节转换成整形
func bytesToInt64(b []byte) (int64, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var x int64
	err := binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x, err
}

func mapToCommand(data []byte, jsonMap map[string]interface{}) (command, error) {
	ctype, ok := jsonMap[CommandCode].(float64)
	if !ok {
		return NewErrorCommand(), errors.New("command is not illegal command type")
	}
	switch ctype {
	case float64(GET):
		var cmd = getCommand{}
		err := json.Unmarshal(data, &cmd)
		return cmd, err
	case float64(SET):
		var cmd = setCommand{}
		err := json.Unmarshal(data, &cmd)
		return cmd, err
	case float64(DEL):
		var cmd = delCommand{}
		err := json.Unmarshal(data, &cmd)
		return cmd, err
	default:
		errorStr := fmt.Sprintf("unknown command type : [%b]", ctype)
		return NewErrorCommand(), errors.New(errorStr)
	}
}
