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
	REQ_CAST = 0x01

	ACK_CAST_OK = 0x00
	ACK_CAST_ROUTINE_NOT_FOUND = 0x01
)

var PORTS = []uint16{
	26130, 26131, 26132, 26133, 26134, 26135, 26136, 26137, 26138, 26139,
}

func Init(name string) {
	nameAndHost := make([]string, 2)
	nameAndHost = strings.SplitN(name, "@", 2)
	agent.name = nameAndHost[0]
	agent.host = nameAndHost[1]
	accept()
	go serve()
}

func accept() {
	var errMessages []string
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
	if len(errMessages) > 0 {
		panic(strings.Join(errMessages, "\n"))
	}
}

func serve() {
	for {
		conn, aErr := agent.lisener.AcceptTCP()
		if aErr != nil {
			// handle accept error
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	for {
		lengthBuffer := make([]byte, 2)
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
		if answer, err := dispatchRequest(code, request); err != nil {
			// handle error
			continue
		} else {
			if _, wErr := conn.Write(answer); wErr != nil {
				// handle error
				continue
			}
		}
	}
}

func dispatchRequest(code byte, request []byte) ([]byte, error) {
	var answer []byte
	var err error
	switch code {
	case REQ_CAST:
		answer, err = handleCast(request)
	default:
		answer, err = []byte{}, errors.New("godist: REQ code error")
	}
	return answer, err
}

/**
 *
 * Connect message described
 * +----------------------------------------------+
 * | port | nameLen | name    | hostLen | host    |
 * |----------------------------------------------|
 * | 2    | 2       | nameLen | 2       | hostLen |
 * +----------------------------------------------+
 */
func handleConnect(request []byte) ([]byte, error) {
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
	registerNode(&base.Node{
		Name: name,
		Host: host,
		Port: port,
	})
	return []byte{}, nil
}

/**
 *
 * Cast message described
 * +----------------------------------------------+
 * | routine id | message length | message        |
 * |------------|----------------|----------------|
 * | 8          | 8              | message length |
 * +----------------------------------------------+
 *
 * Answer message described
 * +--------+
 * | result |
 * |--------|
 * | 1      |
 * +--------+
 */
func handleCast(request []byte) ([]byte, error) {
	routineId := base.RoutineId(binary.LittleEndian.Uint64(request[:8]))
	length := binary.LittleEndian.Uint64(request[8:16])
	message := request[16:16+length]
	if routine, exist := find(routineId); exist {
		routine.Cast(message)
		return []byte{ACK_CAST_OK}, nil
	} else {
		err := errors.New("godist: cast target not found")
		return []byte{ACK_CAST_ROUTINE_NOT_FOUND}, err
	}
}
