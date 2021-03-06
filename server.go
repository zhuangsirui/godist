package godist

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/zhuangsirui/binpacker"
	"github.com/zhuangsirui/godist/base"
)

const (
	REQ_CAST      = 0x01
	REQ_CONN      = 0x02
	REQ_QUERY_ALL = 0x03

	ACK_CONN_OK                = 0x01
	ACK_CONN_NODE_EXIST        = 0x02
	ACK_CAST_OK                = 0x03
	ACK_CAST_ROUTINE_NOT_FOUND = 0x04
	ACK_QUERY_ALL_OK           = 0x05
	ACK_QUERY_ALL_ERR          = 0x06
	ACK_CONN_IS_RETURN         = 0x07
	ACK_CONN_IS_NOT_RETURN     = 0x08
)

var PORTS = []uint16{
	26130, 26131, 26132, 26133, 26134, 26135, 26136, 26137, 26138, 26139,
	9190, 9191, 9192, 9193, 9194, 9195, 9196, 9197, 9198, 9199,
}

// 监听目标端口。
func (agent *Agent) Listen() {
	agent.listen()
	agent.registerNode(agent.Node())
}

func (agent *Agent) listen() {
	var errMessages []string
	agent.listener = nil
	for _, port := range PORTS {
		agent.node.Port = port
		listenStr := fmt.Sprintf("%s:%d", agent.Host(), agent.Port())
		listenAddr, rErr := net.ResolveTCPAddr("tcp", listenStr)
		if rErr != nil {
			errMessages = append(
				errMessages,
				fmt.Sprintf("godist.agent net.ResolveTCPAddr error: %s", rErr),
			)
			continue
		}
		listener, lErr := net.ListenTCP("tcp", listenAddr)
		if lErr != nil {
			errMessages = append(
				errMessages,
				fmt.Sprintf("godist.agent net.ListenTCP error: %s", lErr),
			)
			continue
		}
		agent.listener = listener
		break
	}
	if agent.listener == nil {
		log.Panicf(strings.Join(errMessages, "\n"))
	}
	log.Printf("godist.agent Listen %s successful.", agent.listener.Addr())
}

// 接收请求循环。
func (agent *Agent) Serve() {
	defer func() {
		agent.listener.Close()
		if !agent.isStop {
			agent.listen()
			log.Printf("godist.agent agent restarted")
			go agent.Serve()
		}
	}()
	for {
		conn, err := agent.listener.AcceptTCP()
		if err != nil {
			// handle accept error
			log.Printf("godist.agent accept tcp error: %s", err)
			close(agent.stopped)
			break
		}
		go agent.handleConnection(conn)
	}
}

// Request message described
// +------------------+
// | length | request |
// |------------------|
// | 8      | length  |
// +------------------+
func (agent *Agent) handleConnection(conn *net.TCPConn) {
	for {
		lengthBuffer := make([]byte, 8)
		if _, err := conn.Read(lengthBuffer); err != nil {
			// handle error
			conn.Close()
			break
		}
		length := binary.LittleEndian.Uint16(lengthBuffer)
		buffer := make([]byte, length)
		if _, err := conn.Read(buffer); err != nil {
			// handle error
			continue
		}
		code, request := buffer[0], buffer[1:]
		answer, err := agent.dispatchRequest(code, request)
		if err != nil {
			conn.Close()
			break
		}
		if _, err := conn.Write(answer); err != nil {
			conn.Close()
			break
		}
	}
}

// 分发请求。如果返回 error ，则中断该链接。
func (agent *Agent) dispatchRequest(code byte, request []byte) ([]byte, error) {
	var answer []byte
	var err error
	switch code {
	case REQ_CONN:
		answer, err = agent.handleConnect(request)
	case REQ_CAST:
		answer, err = agent.handleCast(request)
	case REQ_QUERY_ALL:
		answer, err = agent.handleQueryAllNodes(request)
	default:
		answer, err = []byte{}, errors.New("godist: REQ code error")
	}
	return answer, err
}

// Connect message described
// +----------------------------------------------------------+
// | port | nameLen | name    | hostLen | host    | is return |
// |----------------------------------------------|-----------|
// | 2    | 2       | nameLen | 2       | hostLen | 1         |
// +----------------------------------------------------------+
//
// Answer message described
// +--------+
// | result |
// |--------|
// | 1      |
// +--------+
func (agent *Agent) handleConnect(request []byte) ([]byte, error) {
	unpacker := binpacker.NewUnpacker(endian, bytes.NewBuffer(request))
	var isReturn byte
	var port uint16
	var name, host string
	unpacker.FetchByte(&isReturn).
		FetchUint16(&port).
		StringWithUint16Prefix(&name).
		StringWithUint16Prefix(&host)
	node := &base.Node{
		Name: name,
		Host: host,
		Port: port,
	}
	agent.registerNode(node)
	if isReturn != ACK_CONN_IS_RETURN {
		go agent.connectTo(node.FullName(), true)
	}
	return []byte{ACK_CONN_OK}, nil
}

// Query message described
// +-------------------------------------+
// | node name length | node name        |
// |-------------------------------------|
// | 2                | node name length |
// +-------------------------------------+
//
// Answer message described
// +----------------------------------------------------------------------------------------------------------+
// |        |            | first node message                                           | second node message |
// |--------|-------------------------------------------------------------------------------------------------|
// | result | node count | port | name length | name        | host length | host        | ...                 |
// |--------|-------------------------------------------------------------------------------------------------|
// | 1      | 2          | 2    | 2           | name length | 2           | host length | ...                 |
// +----------------------------------------------------------------------------------------------------------+
func (agent *Agent) handleQueryAllNodes(request []byte) ([]byte, error) {
	nameLength := binary.LittleEndian.Uint16(request[:2])
	name := string(request[2 : 2+nameLength])
	if agent.nodeExist(name) {
		agent.nodeLock.RLock()
		defer agent.nodeLock.RUnlock()
		var nodeCount uint16 = 0
		requestBuf := new(bytes.Buffer)
		pk := binpacker.NewPacker(endian, requestBuf).
			PushByte(ACK_QUERY_ALL_OK).
			PushUint16(uint16(len(agent.nodes)))
		for _, node := range agent.nodes {
			pk.PushUint16(node.Port)
			pk.PushUint16(uint16(len(node.Name)))
			pk.PushString(node.Name)
			pk.PushUint16(uint16(len(node.Host)))
			pk.PushString(node.Host)
			nodeCount += 1
		}
		return requestBuf.Bytes(), nil
	}
	return []byte{ACK_QUERY_ALL_ERR}, nil
}

// Cast message described
// +----------------------------------------------+
// | routine id | message length | message        |
// |------------|----------------|----------------|
// | 8          | 8              | message length |
// +----------------------------------------------+
//
// Answer message described
// +--------+
// | result |
// |--------|
// | 1      |
// +--------+
func (agent *Agent) handleCast(request []byte) ([]byte, error) {
	var routineId uint64
	var message []byte
	binpacker.NewUnpacker(endian, bytes.NewBuffer(request)).
		FetchUint64(&routineId).
		BytesWithUint64Perfix(&message)
	if routine, exist := agent.findRoutine(base.RoutineId(routineId)); exist {
		routine.Cast(message)
		return []byte{ACK_CAST_OK}, nil
	} else {
		return []byte{ACK_CAST_ROUTINE_NOT_FOUND}, nil
	}
}
