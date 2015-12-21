package gpmd

import (
	"encoding/binary"
	"errors"
	"godist/base"
	"log"
	"net"
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

func (m *Manager) acceptLoop() {
	for {
		conn, err := m.listener.AcceptTCP()
		if err != nil {
			// TODO handle accept error
			log.Printf("Accept error %s", err)
			break
		}
		// 同步调用，原子性处理各个节点的请求
		log.Printf("Handle connection...")
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
	lengthBuffer := make([]byte, 2)
	if _, err := conn.Read(lengthBuffer); err != nil {
		return err
	}
	length := binary.LittleEndian.Uint16(lengthBuffer)
	requestBuffer := make([]byte, length)
	if _, err := conn.Read(requestBuffer); err != nil {
		return err
	}
	code, request := requestBuffer[0], requestBuffer[1:]
	answer, err := m.dispatchRequest(code, request)
	conn.Write(answer)
	return err
}

/**
 * 在回应之前统一加上 `REQUEST CODE` 。
 */
func (m *Manager) dispatchRequest(code byte, request []byte) ([]byte, error) {
	log.Printf("Code %d request: %v", code, request)
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
	// 1. port
	port := binary.LittleEndian.Uint16(request[:2])
	// 2. name length
	nameLen := binary.LittleEndian.Uint16(request[2:4])
	// 3. name
	name := string(request[4 : 4+nameLen])
	// 4. host length
	hostLen := binary.LittleEndian.Uint16(request[4+nameLen : 4+nameLen+2])
	// 5. host
	host := string(request[4+nameLen+2 : 4+nameLen+2+hostLen])
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
	nameLen := binary.LittleEndian.Uint16(request[:2])
	name := string(request[2 : 2+nameLen])
	node, exist := m.find(name)
	if !exist {
		answer := []byte{ACK_RES_NODE_NOT_EXIST}
		return answer, errors.New("node not exists")
	}
	// 1. Push answer head
	answer := []byte{ACK_RES_OK}
	// 2. Push port number
	portBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBuffer, node.Port)
	answer = append(answer, portBuffer...)
	// 3. Push node name length
	nameLengthBuffer := make([]byte, 2)
	nameLength := len([]byte(node.Name))
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(nameLength))
	answer = append(answer, nameLengthBuffer...)
	// 4. Push node name
	answer = append(answer, []byte(node.Name)...)
	return answer, nil
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
	nameLen := binary.LittleEndian.Uint16(request[:2])
	name := string(request[2 : 2+nameLen])
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
