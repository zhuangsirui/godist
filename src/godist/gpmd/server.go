package gpmd

import(
	"net"
	//"bytes"
	"errors"
	"encoding/binary"
)

const(
	REQ_REGISTER   = 0x01
	REQ_UNREGISTER = 0x02
	REQ_QUERY      = 0x03
)

const(
	ACK_RES_OK             = 0x00
	ACK_RES_NODE_EXIST     = 0x01
	ACK_RES_NODE_NOT_EXIST = 0x02
)

/**
 * Request message described
 * +----------------------------+
 * | length | code | request    |
 * |--------|------|------------|
 * | 2      | 1    | length - 1 |
 * +----------------------------+
 */
func handleRequest(conn *net.TCPConn) error {
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
	answer, err := dispatchRequest(code, request)
	conn.Write(answer)
	return err
}

/**
 * 在回应之前统一加上 `REQUEST CODE` 。
 */
func dispatchRequest(code byte, request []byte) ([]byte, error) {
	var answer []byte
	var err error
	switch code {
	case REQ_REGISTER:
		answer, err = handleRegister(request)
	case REQ_UNREGISTER:
		answer, err = handleUnregister(request)
	case REQ_QUERY:
		answer, err = handleQuery(request)
	default:
		// ignore
	}
	return append([]byte{code}, answer...), err
}

/**
 * Request message described
 * +----------------------------------+
 * | port | name length | name        |
 * |------|-------------|-------------|
 * | 2    | 2           | name length |
 * +----------------------------------+
 *
 * Answer message described
 * +--------+
 * | result |
 * |--------|
 * | 1      |
 * +--------+
 */
func handleRegister(request []byte) ([]byte, error) {
	port := binary.LittleEndian.Uint16(request[:2])
	nameLen := binary.LittleEndian.Uint16(request[2:4])
	name := string(request[5:5+nameLen])
	ok := register(node{
		port: port,
		name: name,
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
func handleQuery(request []byte) ([]byte, error) {
	nameLen := binary.LittleEndian.Uint16(request[:2])
	name := string(request[3:3+nameLen])
	node, exist := find(name)
	if !exist {
		answer := []byte{ACK_RES_NODE_NOT_EXIST}
		return answer, errors.New("node not exists")
	}
	// 1. Push answer head
	answer := []byte{ACK_RES_OK}
	// 2. Push port number
	portBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBuffer, node.port)
	answer = append(answer, portBuffer...)
	// 3. Push node name length
	nameLengthBuffer := make([]byte, 2)
	nameLength := len([]byte(node.name))
	binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(nameLength))
	answer = append(answer, nameLengthBuffer...)
	// 4. Push node name
	answer = append(answer, []byte(node.name)...)
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
func handleUnregister(request []byte) ([]byte, error) {
	nameLen := binary.LittleEndian.Uint16(request[:2])
	name := string(request[3:3+nameLen])
	ok := unregister(name)
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