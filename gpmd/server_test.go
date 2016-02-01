package gpmd

import (
	"bytes"
	"fmt"
	"godist/binary/packer"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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
	requestBuf := new(bytes.Buffer)
	packer.NewPacker(requestBuf).
		PushByte(REQ_REGISTER).
		PushUint16(20130).
		PushUint16(uint16(len("agent_name"))).
		PushString("agent_name").
		PushUint16(uint16(len(testHost))).
		PushString(testHost)
	request := packer.AddUint16Perfix(requestBuf.Bytes())
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	assert.Equal(t, err, nil, "ResolveTCPAddr Error.")
	conn, err := net.DialTCP("tcp", nil, address)
	assert.Equal(t, err, nil, "DialTCP Error.")
	_, err = conn.Write(request)
	assert.Equal(t, err, nil, "Write Error.")
	ack := make([]byte, 2)
	_, err = conn.Read(ack)
	assert.Equal(t, err, nil, "Read Error.")
	assert.Equal(t, ack, []byte{REQ_REGISTER, ACK_RES_OK}, "Ack Error.")
}

func TestRegisterAnExistNode(t *testing.T) {
	requestBuf := new(bytes.Buffer)
	packer.NewPacker(requestBuf).
		PushByte(REQ_REGISTER).
		PushUint16(26130).
		PushUint16(uint16(len("agent_name"))).
		PushString("agent_name").
		PushUint16(uint16(len(testHost))).
		PushString(testHost)
	request := packer.AddUint16Perfix(requestBuf.Bytes())
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	assert.Equal(t, err, nil, "ResolveTCPAddr Error.")
	conn, err := net.DialTCP("tcp", nil, address)
	assert.Equal(t, err, nil, "DialTCP Error.")
	_, err = conn.Write(request)
	assert.Equal(t, err, nil, "Write Error.")
	ack := make([]byte, 2)
	_, err = conn.Read(ack)
	assert.Equal(t, err, nil, "Read Error.")
	assert.Equal(t, ack, []byte{REQ_REGISTER, ACK_RES_NODE_EXIST}, "Ack Error.")
}

func TestQuery(t *testing.T) {
	requestBuf := new(bytes.Buffer)
	packer.NewPacker(requestBuf).
		PushByte(REQ_QUERY).
		PushUint16(uint16(len("agent_name"))).
		PushString("agent_name")
	request := packer.AddUint16Perfix(requestBuf.Bytes())
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	assert.Equal(t, err, nil, "ResolveTCPAddr Error.")
	conn, err := net.DialTCP("tcp", nil, address)
	assert.Equal(t, err, nil, "DialTCP Error.")
	_, err = conn.Write(request)
	assert.Equal(t, err, nil, "Write Error.")
	ack := make([]byte, 2)
	_, err = conn.Read(ack)
	assert.Equal(t, err, nil, "Read Error.")
	assert.Equal(t, ack, []byte{REQ_QUERY, ACK_RES_OK}, "Ack Error.")
}

func TestQueryNotExist(t *testing.T) {
	requestBuf := new(bytes.Buffer)
	packer.NewPacker(requestBuf).
		PushByte(REQ_QUERY).
		PushUint16(uint16(len("fake_name"))).
		PushString("fake_name")
	request := packer.AddUint16Perfix(requestBuf.Bytes())
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	assert.Equal(t, err, nil, "ResolveTCPAddr Error.")
	conn, err := net.DialTCP("tcp", nil, address)
	assert.Equal(t, err, nil, "DialTCP Error.")
	_, err = conn.Write(request)
	assert.Equal(t, err, nil, "Write Error.")
	ack := make([]byte, 2)
	_, err = conn.Read(ack)
	assert.Equal(t, err, nil, "Read Error.")
	assert.Equal(t, ack, []byte{REQ_QUERY, ACK_RES_NODE_NOT_EXIST}, "Ack Error.")
}

func TestUnregister(t *testing.T) {
	requestBuf := new(bytes.Buffer)
	packer.NewPacker(requestBuf).
		PushByte(REQ_UNREGISTER).
		PushUint16(uint16(len("agent_name"))).
		PushString("agent_name")
	request := packer.AddUint16Perfix(requestBuf.Bytes())
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	assert.Equal(t, err, nil, "ResolveTCPAddr Error.")
	conn, err := net.DialTCP("tcp", nil, address)
	assert.Equal(t, err, nil, "DialTCP Error.")
	_, err = conn.Write(request)
	assert.Equal(t, err, nil, "Write Error.")
	ack := make([]byte, 2)
	_, err = conn.Read(ack)
	assert.Equal(t, err, nil, "Read Error.")
	assert.Equal(t, ack, []byte{REQ_UNREGISTER, ACK_RES_OK}, "Ack Error.")
}

func TestUnregister2(t *testing.T) {
	requestBuf := new(bytes.Buffer)
	packer.NewPacker(requestBuf).
		PushByte(REQ_UNREGISTER).
		PushUint16(uint16(len("agent_name"))).
		PushString("agent_name")
	request := packer.AddUint16Perfix(requestBuf.Bytes())
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
	assert.Equal(t, err, nil, "ResolveTCPAddr Error.")
	conn, err := net.DialTCP("tcp", nil, address)
	assert.Equal(t, err, nil, "DialTCP Error.")
	_, err = conn.Write(request)
	assert.Equal(t, err, nil, "Write Error.")
	ack := make([]byte, 2)
	_, err = conn.Read(ack)
	assert.Equal(t, err, nil, "Read Error.")
	assert.Equal(t, ack, []byte{REQ_UNREGISTER, ACK_RES_NODE_NOT_EXIST}, "Ack Error.")
}

func TestStop(t *testing.T) {
	Stop()
}
