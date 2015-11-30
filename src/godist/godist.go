package godist

import(
	"net"
	//"errors"
	"encoding/binary"
	"fmt"
	"strings"
	"sync/atomic"
	"godist/base"
	"godist/gpmd"
)

const EPMD_PORT = 2613

var routineCounter uint64

// Agent 结构持有本节点所有注册过的 Goroutine 对象，所有在集群中的节点信息以及
// 针对所有节点的链接。
type Agent struct {
	name string
	host string
	port uint16
	routineCounter *uint64
	gpmd base.GPMD
	lisener *net.TCPListener
	nodes map[string]*base.Node
	routines map[base.RoutineId]*base.Routine
	connections map[string]*net.TCPConn
}

// 构建 godist.Agent 对象，返回其指针。
func New(name string) *Agent {
	nameAndHost := make([]string, 2)
	nameAndHost = strings.SplitN(name, "@", 2)
	return &Agent{
		routineCounter: &routineCounter,
		name: nameAndHost[0],
		host: nameAndHost[1],
		gpmd: base.GPMD{
			Host: "",
			Port: EPMD_PORT,
		},
		nodes: make(map[string]*base.Node),
		routines: make(map[base.RoutineId]*base.Routine),
		connections: make(map[string]*net.TCPConn),
	}
}

// 设置本机的 GPMD 服务地址。默认为 ":2613"
func (agent *Agent) SetGPMD(host string, port uint16) {
	agent.gpmd.Host = host
	agent.gpmd.Port = port
}

// 向本地的 agent 注册一个 Goroutine 。如果该 Goroutine 对象已经被设置过 Id ，则
// 会抛出 panic 。
func (agent *Agent) RegisterRoutine(routine *base.Routine) {
	routine.SetId(agent.incrRoutineId())
	agent.registerRoutine(routine)
}

// 向本地 GPMD 注册节点信息，无法注册会 panic 异常。
func (agent *Agent) Register() {
	resolvedAddr, rErr := net.ResolveTCPAddr("tcp", agent.gpmd.Address())
	if rErr != nil {
		panic(fmt.Sprintf("godist: GPMD address error: %s", rErr))
	}
	conn, dErr := net.DialTCP("tcp", nil, resolvedAddr)
	if dErr != nil {
		panic(fmt.Sprintf("godist: GPMD dial error: %s", dErr))
	}
	request := []byte{gpmd.REQ_REGISTER}
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
	requestLengthBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
	request = append(requestLengthBuffer, request...)
	if _, wErr := conn.Write(request); wErr != nil {
		panic(fmt.Sprintf("godist: Send register message error: %s", wErr))
	}
	ackBuffer := make([]byte, 2)
	if _, rErr := conn.Read(ackBuffer); rErr != nil {
		panic(fmt.Sprintf("godist: Receive register message error: %s", rErr))
	}
	if ackBuffer[0] != gpmd.REQ_REGISTER || ackBuffer[1] != gpmd.ACK_RES_OK {
		panic(fmt.Sprintf("godist: Register failed. %v", ackBuffer))
	}
}

// 向目标 Goroutine 发送消息。
func (agent *Agent) CastTo(nodeName string, routineId base.RoutineId, message []byte) {
	if conn, exist := agent.connections[nodeName]; exist {
		request := []byte{REQ_CAST}
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

func (agent *Agent) QueryAllNode(nodeName string) {
	name, _ := parseNameAndHost(nodeName)
	if conn, exist := agent.connections[name]; exist {
		// REQUEST
		request := []byte{REQ_QUERY_ALL}
		// 1. name length
		nameLengthBuffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len(agent.name)))
		request = append(request, nameLengthBuffer...)
		// 2. name
		request = append(request, []byte(agent.name)...)
		// 3. add length prefix
		requestLengthBuffer := make([]byte, 8)
		binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
		request = append(requestLengthBuffer, request...)
		if _, wErr := conn.Write(request); wErr != nil {
			return
		}
		// ANSWER
		ackCodeBuf := make([]byte, 1)
		if _, rErr := conn.Read(ackCodeBuf); rErr != nil {
			return
		}
		if ackCodeBuf[0] != ACK_QUERY_ALL_OK {
			return
		}
		lengthBuf := make([]byte, 2)
		if _, rLErr := conn.Read(lengthBuf); rLErr != nil {
			return
		}
		length := binary.LittleEndian.Uint16(lengthBuf)
		answer := make([]byte, length)
		if _, rAErr := conn.Read(answer); rAErr != nil {
			return
		}
		countBuf := make([]byte, 2)
		countBuf, answer = answer[:2], answer[2:]
		count := int(binary.LittleEndian.Uint16(countBuf))
		for i := 0; i < count; i++ {
			var portBuf, nameLenBuf, nameBuf, hostLenBuf, hostBuf []byte
			portBuf, answer = answer[:2], answer[2:]
			port := binary.LittleEndian.Uint16(portBuf)
			nameLenBuf, answer = answer[:2], answer[2:]
			nameLen := binary.LittleEndian.Uint16(nameLenBuf)
			nameBuf, answer = answer[:nameLen], answer[nameLen:]
			name := string(nameBuf)
			hostLenBuf, answer = answer[:2], answer[2:]
			hostLen := binary.LittleEndian.Uint16(hostLenBuf)
			hostBuf, answer = answer[:hostLen], answer[hostLen:]
			host := string(hostBuf)
			node := &base.Node{
				Port: port,
				Host: host,
				Name: name,
			}
			agent.registerNode(node)
			if !agent.connExist(name) {
				go func() {
					agent.ConnectTo(node.FullName())
				}()
			}
		}
	}
}

// 查询目标节点的详细信息。
func (agent *Agent) QueryNode(nodeName string) {
	name, host := parseNameAndHost(nodeName)
	if !agent.nodeExist(name) {
		gpmdAddr, rErr := net.ResolveTCPAddr("tcp", fmt.Sprintf(
			"%s:%d", host, agent.gpmd.Port,
		))
		if rErr != nil {
			return
		}
		conn, dErr := net.DialTCP("tcp", nil, gpmdAddr)
		defer conn.Close()
		if dErr != nil {
			return
		}
		request := []byte{gpmd.REQ_QUERY}
		// 1. name length
		nameLengthBuffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(nameLengthBuffer, uint16(len(name)))
		request = append(request, nameLengthBuffer...)
		// 2. name
		request = append(request, []byte(name)...)
		// 3. add length prefix
		requestLengthBuffer := make([]byte, 2)
		binary.LittleEndian.PutUint16(requestLengthBuffer, uint16(len(request)))
		request = append(requestLengthBuffer, request...)
		if _, wErr := conn.Write(request); wErr != nil {
			return
		}
		resultCodeBuffer := make([]byte, 2)
		if _, rCErr1 := conn.Read(resultCodeBuffer); rCErr1 != nil {
			return
		}
		if resultCodeBuffer[1] != gpmd.ACK_RES_OK {
			return
		}
		portBuffer := make([]byte, 2)
		if _, rPErr := conn.Read(portBuffer); rPErr != nil {
			return
		}
		port := binary.LittleEndian.Uint16(portBuffer)
		ackNameLengthBuffer := make([]byte, 2)
		if _, rNLErr := conn.Read(ackNameLengthBuffer); rNLErr != nil {
			return
		}
		nameLength := binary.LittleEndian.Uint16(ackNameLengthBuffer)
		nameBuffer := make([]byte, nameLength)
		if _, rNErr := conn.Read(nameBuffer); rNErr != nil {
			return
		}
		ackName := string(nameBuffer)
		if ackName != name {
			return
		}
		agent.registerNode(&base.Node{
			Port: port,
			Host: host,
			Name: name,
		})
	}
}

// 尝试向另一个节点建立连接。建立好之后会一直保持持有连接。用于节点之间的
// Goroutine 消息收发。
func (agent *Agent) ConnectTo(nodeName string) {
	name, _ := parseNameAndHost(nodeName)
	if node, exist := agent.nodes[name]; exist {
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
		request := []byte{REQ_CONN}
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
		agent.connections[name] = conn
	}
}

func (agent *Agent) find(routineId base.RoutineId) (*base.Routine, bool) {
	routine, exist := agent.routines[routineId]
	return routine, exist
}

func (agent *Agent) nodeExist(name string) bool {
	_, exist := agent.nodes[name]
	return exist
}

func (agent *Agent) connExist(name string) bool {
	_, exist := agent.connections[name]
	return exist
}

func (agent *Agent) registerNode(node *base.Node) {
	if _, exist := agent.nodes[node.Name]; !exist {
		agent.nodes[node.Name] = node
	}
}

func (agent *Agent) registerRoutine(routine *base.Routine) {
	if _, exist := agent.routines[routine.GetId()]; !exist {
		agent.routines[routine.GetId()] = routine
	}
}

func (agent *Agent) incrRoutineId() base.RoutineId {
	id := atomic.AddUint64(agent.routineCounter, 1)
	return base.RoutineId(id)
}

func parseNameAndHost(nodeName string) (string, string) {
	nameAndHost := make([]string, 2)
	nameAndHost = strings.SplitN(nodeName, "@", 2)
	name := nameAndHost[0]
	host := nameAndHost[1]
	return name, host
}
