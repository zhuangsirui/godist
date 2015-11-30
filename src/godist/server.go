package godist

import(
	"fmt"
	"net"
	"errors"
	"encoding/binary"
	"strings"
	"godist/base"
)

const(
	REQ_CAST      = 0x01
	REQ_CONN      = 0x02
	REQ_QUERY_ALL = 0x03

	ACK_CONN_OK                = 0x01
	ACK_CONN_NODE_EXIST        = 0x02
	ACK_CAST_OK                = 0x03
	ACK_CAST_ROUTINE_NOT_FOUND = 0x04
	ACK_QUERY_ALL_OK           = 0x05
	ACK_QUERY_ALL_ERR          = 0x06
)

var PORTS = []uint16{
	26130, 26131, 26132, 26133, 26134, 26135, 26136, 26137, 26138, 26139,
}

// 监听目标端口。
func (agent *Agent) Listen() {
	var errMessages []string
	agent.lisener = nil
	for _, port := range PORTS {
		agent.port = port
		listenStr := fmt.Sprintf("%s:%d", agent.host, agent.port)
		listenAddr, rErr := net.ResolveTCPAddr("tcp", listenStr)
		if rErr != nil {
			errMessages = append(
				errMessages,
				fmt.Sprintf("godist.agent net.ResolveTCPAddr error: %s", rErr),
			)
			continue
		}
		lisener, lErr := net.ListenTCP("tcp", listenAddr)
		if lErr != nil {
			errMessages = append(
				errMessages,
				fmt.Sprintf("godist.agent net.ListenTCP error: %s", lErr),
			)
			continue
		}
		agent.lisener = lisener
		break
	}
	if agent.lisener == nil {
		panic(strings.Join(errMessages, "\n"))
	}
}

// 接收请求循环。
func (agent *Agent) Serve() {
	for {
		conn, aErr := agent.lisener.AcceptTCP()
		if aErr != nil {
			// handle accept error
			continue
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
			continue
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
// +----------------------------------------------+
// | port | nameLen | name    | hostLen | host    |
// |----------------------------------------------|
// | 2    | 2       | nameLen | 2       | hostLen |
// +----------------------------------------------+
//
// Answer message described
// +--------+
// | result |
// |--------|
// | 1      |
// +--------+
func (agent *Agent) handleConnect(request []byte) ([]byte, error) {
	// 1. port
	port := binary.LittleEndian.Uint16(request[:2])
	// 2. name length
	nLength := binary.LittleEndian.Uint16(request[2:4])
	// 3. name
	name := string(request[4:4+nLength])
	// 4. host length
	hLength := binary.LittleEndian.Uint16(request[4+nLength:4+nLength+2])
	// 5. host name
	host := string(request[4+nLength+2:4+nLength+2+hLength])
	node := &base.Node{
		Name: name,
		Host: host,
		Port: port,
	}
	agent.registerNode(node)
	if !agent.connExist(name) {
		go agent.ConnectTo(node.FullName())
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
// +-------------------------------------------------------------------------------------------------------------------+
// |        |        |            | first node message                                           | second node message |
// |--------|--------|-------------------------------------------------------------------------------------------------|
// | result | length | node count | port | name length | name        | host length | host        | ...                 |
// |--------|--------|-------------------------------------------------------------------------------------------------|
// | 1      | 2      | 2          | 2    | 2           | name length | 2           | host length | ...                 |
// +-------------------------------------------------------------------------------------------------------------------+
func (agent *Agent) handleQueryAllNodes(request []byte) ([]byte, error) {
	nameLength := binary.LittleEndian.Uint16(request[:2])
	name := string(request[2:2+nameLength])
	if agent.nodeExist(name) {
		var nodeCount uint16 = 0
		//answer := []byte{ACK_QUERY_ALL_OK}
		answer := []byte{}
		for _, node := range agent.nodes {
			// 1. port
			portBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(portBuf, node.Port)
			answer = append(answer, portBuf...)
			// 2. name length
			nameLenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(nameLenBuf, uint16(len(node.Name)))
			answer = append(answer, nameLenBuf...)
			// 3. name
			answer = append(answer, []byte(node.Name)...)
			// 4. host length
			hostLenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(hostLenBuf, uint16(len(node.Host)))
			answer = append(answer, hostLenBuf...)
			// 5. host
			answer = append(answer, []byte(node.Host)...)
			nodeCount += 1
		}
		countBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(countBuf, nodeCount)
		answer = append(countBuf, answer...)
		lengthBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lengthBuf, uint16(len(answer)))
		answer = append(lengthBuf, answer...)
		answer = append([]byte{ACK_QUERY_ALL_OK}, answer...)
		return answer, nil
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
	routineId := base.RoutineId(binary.LittleEndian.Uint64(request[:8]))
	length := binary.LittleEndian.Uint64(request[8:16])
	message := request[16:16+length]
	if routine, exist := agent.find(routineId); exist {
		routine.Cast(message)
		return []byte{ACK_CAST_OK}, nil
	} else {
		return []byte{ACK_CAST_ROUTINE_NOT_FOUND}, nil
	}
}
