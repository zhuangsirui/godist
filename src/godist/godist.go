package godist

import(
	"net"
	//"errors"
	"encoding/binary"
	"fmt"
	"godist/base"
)

// Agent struct hold all routines infomation on itself go process and all the
// other nodes' host, name and port.
type Agent struct {
	name string
	host string
	port uint16
	lisener *net.TCPListener
	nodes map[string]*base.Node
	routines map[base.RoutineId]*base.Routine
	connections map[string]*net.TCPConn
}

var agent = &Agent{
	nodes: make(map[string]*base.Node),
	routines: make(map[base.RoutineId]*base.Routine),
}

func find(routineId base.RoutineId) (*base.Routine, bool) {
	routine, exist := agent.routines[routineId]
	return routine, exist
}

func nodeExist(name string) bool {
	_, exist := agent.nodes[name]
	return exist
}

func registerNode(node *base.Node) {
	if _, exist := agent.nodes[node.Name]; !exist {
		agent.nodes[node.Name] = node
	}
}

func ConnectTo(nodeName string) {
	if node, exist := agent.nodes[nodeName]; exist {
		address, rErr := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", node.Host, node.Port))
		if rErr != nil {
			// handle error
			return
		}
		conn, dErr := net.DialTCP("tcp", nil, address)
		if dErr != nil {
			// handle error
			conn.Close()
			return
		}
		request := []byte{}
		// 1. port
		portBuffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(portBuffer, agent.port)
		request = append(request, portBuffer...)
		// 2. name length
		nameLengthBuffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len(agent.name)))
		request = append(request, nameLengthBuffer...)
		// 3. name
		request = append(request, []byte(agent.name)...)
		// 4. host length
		hostLengthBuffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(hostLengthBuffer, uint16(len(agent.host)))
		request = append(request, hostLengthBuffer...)
		// 5. host
		request = append(request, []byte(agent.host)...)
		// 6. add length prefix
		requestLengthBuffer := make([]byte, 8)
		binary.LittleEndian.PutUint64(requestLengthBuffer, uint64(len(request)))
		request = append(requestLengthBuffer, request...)
		if _, wErr := conn.Write(request); wErr != nil {
			// handle error
			conn.Close()
			return
		}
		ackBuffer := make([]byte, 1)
		if _, rErr := conn.Read(ackBuffer); rErr != nil {
			conn.Close()
			return
		}
		if ackBuffer[0] != ACK_CONN_OK {
			conn.Close()
			return
		}
		agent.connections[nodeName] = conn
	}
}

func CastTo(nodeName string, routineId base.RoutineId, message []byte) {
	if conn, exist := agent.connections[nodeName]; exist {
		request := []byte{}
		// 1. routine id
		routineIdBuffer := make([]byte, 8)
		binary.LittleEndian.PutUint64(routineIdBuffer, uint64(routineId))
		request = append(request, routineIdBuffer...)
		// 2. message length
		messageLengthBuffer := make([]byte, 8)
		binary.LittleEndian.PutUint64(messageLengthBuffer, uint64(len(message)))
		request = append(request, messageLengthBuffer...)
		// 3. message
		request = append(request, message...)
		// 4. add length prefix
		requestLengthBuffer := make([]byte, 8)
		binary.LittleEndian.PutUint64(requestLengthBuffer, uint64(len(request)))
		request = append(requestLengthBuffer, request...)
		if _, wErr := conn.Write(request); wErr != nil {
			return
		}
		ackBuffer := make([]byte, 1)
		if _, rErr := conn.Read(ackBuffer); rErr != nil {
			return
		}
		if ackBuffer[0] != ACK_CAST_OK {
			return
		}
	}
}
