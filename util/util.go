package util

import (
	"SecShell/config"
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
)

func Packaging(message []byte) ([]byte, error) {
	// message length
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	// write length to buffer
	err := binary.Write(pkg, binary.BigEndian, length)
	if err != nil {
		return nil, err
	}
	// write message in buffer
	err = binary.Write(pkg, binary.BigEndian, message)
	if err != nil {
		return nil, err
	}
	// return buffer as bytes
	return pkg.Bytes(), nil
	//return message,nil
}

func UnPackagingInfo(conn *kcp.UDPSession) config.SystemInfo {
	reader := bufio.NewReader(conn)
	peek, err := reader.Peek(4)
	if err != nil {
		return config.SystemInfo{}
	}
	buffer := bytes.NewBuffer(peek)
	//读取数据长度
	var length int32
	err = binary.Read(buffer, binary.BigEndian, &length)
	if err != nil {
		fmt.Println(err)
		return config.SystemInfo{}
	}
	data := make([]byte, length+4)
	_, err = reader.Read(data)
	if err != nil {
		return config.SystemInfo{}
	}
	//return data[4:]
	//log.Println(string(data[4:]))
	var systems config.SystemInfo
	err = json.Unmarshal(data[4:], &systems)
	//fmt.Println(err)
	if err != nil {
		return config.SystemInfo{}
	}
	return systems
}

func UnPackaging(conn *kcp.UDPSession) []byte {
	reader := bufio.NewReader(conn)
	peek, err := reader.Peek(4)
	if err != nil {
		return nil
	}
	buffer := bytes.NewBuffer(peek)
	//读取数据长度
	var length int32
	err = binary.Read(buffer, binary.BigEndian, &length)
	if err != nil {
		return nil
	}
	data := make([]byte, length+4)
	_, err = reader.Read(data)
	if err != nil {
		return nil
	}
	return data[4:]
}
