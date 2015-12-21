package gpmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"testing"
)

const (
	testPort = 1989
	testHost = "localhost"
)

func TestInitErr(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Now you should panic!")
		}
	}()
	SetPort(22)
	SetHost("google.com")
	Init()
}

func TestInit(t *testing.T) {
	SetPort(testPort)
	SetHost(testHost)
	Init()
}

func TestRegister(t *testing.T) {
	request := []byte{REQ_REGISTER}
	// 1. port
	portBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBuffer, 26130)
	request = append(request, portBuffer...)
	// 2. name length
	nameLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len("agent_name")))
	request = append(request, nameLengthBuffer...)
	// 3. name
	request = append(request, []byte("agent_name")...)
	// 4. host length
	hostLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(hostLengthBuffer, uint16(len(testHost)))
	request = append(request, hostLengthBuffer...)
	// 5. host
	request = append(request, []byte(testHost)...)
	// 6. add length prefix
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	if err != nil {
		t.Error(err)
	}
	conn, dErr := net.DialTCP("tcp", nil, address)
	if dErr != nil {
		t.Error(dErr)
	}
	_, wErr := conn.Write(request)
	if wErr != nil {
		t.Error(dErr)
	}
	ackBuffer := make([]byte, 2)
	_, rErr := conn.Read(ackBuffer)
	if rErr != nil {
		t.Error(rErr)
	}
	if bytes.Compare(ackBuffer, []byte{REQ_REGISTER, ACK_RES_OK}) != 0 {
		t.Error("ack error")
	}
}

func TestRegisterAnExistNode(t *testing.T) {
	request := []byte{REQ_REGISTER}
	// 1. port
	portBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBuffer, 26130)
	request = append(request, portBuffer...)
	// 2. name length
	nameLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len("agent_name")))
	request = append(request, nameLengthBuffer...)
	// 3. name
	request = append(request, []byte("agent_name")...)
	// 4. host length
	hostLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(hostLengthBuffer, uint16(len(testHost)))
	request = append(request, hostLengthBuffer...)
	// 5. host
	request = append(request, []byte(testHost)...)
	// 6. add length prefix
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	if err != nil {
		t.Error(err)
	}
	conn, dErr := net.DialTCP("tcp", nil, address)
	if dErr != nil {
		t.Error(dErr)
	}
	_, wErr := conn.Write(request)
	if wErr != nil {
		t.Error(dErr)
	}
	ackBuffer := make([]byte, 2)
	_, rErr := conn.Read(ackBuffer)
	if rErr != nil {
		t.Error(rErr)
	}
	if bytes.Compare(ackBuffer, []byte{REQ_REGISTER, ACK_RES_NODE_EXIST}) != 0 {
		t.Error("ack error")
	}
}

func TestQuery(t *testing.T) {
	request := []byte{REQ_QUERY}
	// 1. name length
	nameLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len("agent_name")))
	request = append(request, nameLengthBuffer...)
	// 2. name
	request = append(request, []byte("agent_name")...)
	// 3. add length prefix
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	if err != nil {
		t.Error(err)
	}
	conn, dErr := net.DialTCP("tcp", nil, address)
	if dErr != nil {
		t.Error(dErr)
	}
	_, wErr := conn.Write(request)
	if wErr != nil {
		t.Error(dErr)
	}
	ackBuffer := make([]byte, 2)
	_, rErr := conn.Read(ackBuffer)
	if rErr != nil {
		t.Error(rErr)
	}
	if bytes.Compare(ackBuffer, []byte{REQ_QUERY, ACK_RES_OK}) != 0 {
		t.Error("ack error")
	}
}

func TestQueryNotExist(t *testing.T) {
	request := []byte{REQ_QUERY}
	// 1. name length
	nameLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len("not_exist_name")))
	request = append(request, nameLengthBuffer...)
	// 2. name
	request = append(request, []byte("not_exist_name")...)
	// 3. add length prefix
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	if err != nil {
		t.Error(err)
	}
	conn, dErr := net.DialTCP("tcp", nil, address)
	if dErr != nil {
		t.Error(dErr)
	}
	_, wErr := conn.Write(request)
	if wErr != nil {
		t.Error(dErr)
	}
	ackBuffer := make([]byte, 2)
	_, rErr := conn.Read(ackBuffer)
	if rErr != nil {
		t.Error(rErr)
	}
	if bytes.Compare(ackBuffer, []byte{REQ_QUERY, ACK_RES_NODE_NOT_EXIST}) != 0 {
		t.Error("ack error")
	}
}

func TestUnregister(t *testing.T) {
	request := []byte{REQ_UNREGISTER}
	// 1. name length
	nameLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len("agent_name")))
	request = append(request, nameLengthBuffer...)
	// 2. name
	request = append(request, []byte("agent_name")...)
	// 3. add length prefix
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	if err != nil {
		t.Error(err)
	}
	conn, dErr := net.DialTCP("tcp", nil, address)
	if dErr != nil {
		t.Error(dErr)
	}
	_, wErr := conn.Write(request)
	if wErr != nil {
		t.Error(dErr)
	}
	ackBuffer := make([]byte, 2)
	_, rErr := conn.Read(ackBuffer)
	if rErr != nil {
		t.Error(rErr)
	}
	if bytes.Compare(ackBuffer, []byte{REQ_UNREGISTER, ACK_RES_OK}) != 0 {
		t.Error("ack error")
	}
}

func TestUnregister2(t *testing.T) {
	request := []byte{REQ_UNREGISTER}
	// 1. name length
	nameLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len("agent_name")))
	request = append(request, nameLengthBuffer...)
	// 2. name
	request = append(request, []byte("agent_name")...)
	// 6. add length prefix
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	if err != nil {
		t.Error(err)
	}
	conn, dErr := net.DialTCP("tcp", nil, address)
	if dErr != nil {
		t.Error(dErr)
	}
	_, wErr := conn.Write(request)
	if wErr != nil {
		t.Error(dErr)
	}
	ackBuffer := make([]byte, 2)
	_, rErr := conn.Read(ackBuffer)
	if rErr != nil {
		t.Error(rErr)
	}
	if bytes.Compare(ackBuffer, []byte{REQ_UNREGISTER, ACK_RES_NODE_NOT_EXIST}) != 0 {
		t.Error("ack error")
	}
}

func TestStop(t *testing.T) {
	Stop()
}
