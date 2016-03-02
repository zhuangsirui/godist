package gpmd

import (
	"bytes"
	"errors"
	"godist/base"
	"log"
	"net"

	"github.com/zhuangsirui/binpacker"
)

const (
	REQ_REGISTER   = 0x01
	REQ_UNREGISTER = 0x02
	REQ_QUERY      = 0x03
)

const (
	ACK_RES_OK             = 0x00
	ACK_RES_NODE_EXIST     = 0x01
	ACK_RES_NODE_NOT_EXIST = 0x02
)

func (m *Manager) Stop() {
	m.isStop = true
	m.listener.Close()
}

func (m *Manager) acceptLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GPMD accept loop recover error: %s", r)
		}
		if !m.isStop {
			log.Printf("GPMD accept loop is not stop. Reserving...")
			m.listener.Close()
			go m.Serve()
			m.restarted <- true
		} else {
			close(m.stopped)
		}
	}()
	for {
		conn, err := m.listener.AcceptTCP()
		if err != nil {
			// TODO handle accept error
			log.Printf("Accept error %s", err)
			break
		}
		// 同步调用，原子性处理各个节点的请求
		m.handleConnection(conn)
		conn.Close()
	}
}

/**
 * Request message described
 * +----------------------------+
 * | length | code | request    |
 * |--------|------|------------|
 * | 2      | 1    | length - 1 |
 * +----------------------------+
 */
func (m *Manager) handleConnection(conn *net.TCPConn) error {
	unpacker := binpacker.NewUnpacker(conn)
	var requestBuffer []byte
	unpacker.BytesWithUint16Perfix(&requestBuffer)
	code, request := requestBuffer[0], requestBuffer[1:]
	answer, err := m.dispatchRequest(code, request)
	conn.Write(answer)
	return err
}

/**
 * 在回应之前统一加上 `REQUEST CODE` 。
 */
func (m *Manager) dispatchRequest(code byte, request []byte) ([]byte, error) {
	var answer []byte
	var err error
	switch code {
	case REQ_REGISTER:
		answer, err = m.handleRegister(request)
	case REQ_UNREGISTER:
		answer, err = m.handleUnregister(request)
	case REQ_QUERY:
		answer, err = m.handleQuery(request)
	default:
		// ignore
	}
	return append([]byte{code}, answer...), err
}

/**
 * Request message described
 * +--------------------------------------------------+
 * | port | nameLen | name        | hostLen | host    |
 * |------|---------|---------------------------------|
 * | 2    | 2       | nameLen     | 2       | hostLen |
 * +--------------------------------------------------+
 *
 * Answer message described
 * +--------+
 * | result |
 * |--------|
 * | 1      |
 * +--------+
 */
func (m *Manager) handleRegister(request []byte) ([]byte, error) {
	unpacker := binpacker.NewUnpacker(bytes.NewBuffer(request))
	var port uint16
	var name, host string
	unpacker.FetchUint16(&port).
		StringWithUint16Perfix(&name).
		StringWithUint16Perfix(&host)
	ok := m.register(&base.Node{
		Port: port,
		Host: host,
		Name: name,
	})
	var answer []byte
	var err error
	if ok {
		answer = []byte{ACK_RES_OK}
	} else {
		answer = []byte{ACK_RES_NODE_EXIST}
		err = errors.New("node already exists")
	}
	return answer, err
}

/**
 *
 * Request message described
 * +-------------|-------------+
 * | name length | name        |
 * |-------------|-------------|
 * | 2           | name length |
 * +-------------|-------------+
 *
 * Answer message described without node info
 * +--------+
 * | result |
 * |--------|
 * | 1      |
 * +--------+
 *
 * Answer message described with node info
 * +-------------------------------------------+
 * | result | port | name length | node name   |
 * |--------|------|-------------|-------------|
 * | 1      | 2    | 2           | name length |
 * +-------------------------------------------+
 */
func (m *Manager) handleQuery(request []byte) ([]byte, error) {
	unpacker := binpacker.NewUnpacker(bytes.NewBuffer(request))
	var name string
	unpacker.StringWithUint16Perfix(&name)
	node, exist := m.find(name)
	if !exist {
		answer := []byte{ACK_RES_NODE_NOT_EXIST}
		return answer, errors.New("node not exists")
	}
	requestBuf := new(bytes.Buffer)
	binpacker.NewPacker(requestBuf).
		PushByte(ACK_RES_OK).
		PushUint16(node.Port).
		PushUint16(uint16(len(node.Name))).
		PushString(node.Name)
	return requestBuf.Bytes(), nil
}

/**
 *
 * Request message described
 * +-------------|-------------+
 * | name length | name        |
 * |-------------|-------------|
 * | 2           | name length |
 * +-------------|-------------+
 *
 * Answer message described
 * +--------+
 * | result |
 * |--------|
 * | 1      |
 * +--------+
 */
func (m *Manager) handleUnregister(request []byte) ([]byte, error) {
	unpacker := binpacker.NewUnpacker(bytes.NewBuffer(request))
	var name string
	unpacker.StringWithUint16Perfix(&name)
	ok := m.unregister(name)
	var answer []byte
	var err error
	if ok {
		answer = []byte{ACK_RES_OK}
	} else {
		answer = []byte{ACK_RES_NODE_NOT_EXIST}
		err = errors.New("node not exists")
	}
	return answer, err
}
