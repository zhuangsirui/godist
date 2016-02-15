package godist

import (
	"bytes"
	"fmt"
	"godist/base"
	"godist/gpmd"
	"log"
	"net"
	"strings"
	"sync/atomic"

	"github.com/zhuangsirui/binpacker"
)

const EPMD_PORT = 2613

var routineCounter uint64

// Agent 结构持有本节点所有注册过的 Goroutine 对象，所有在集群中的节点信息以及
// 针对所有节点的链接。
type Agent struct {
	gpmd           base.GPMD
	name           string
	host           string
	port           uint16
	nodes          map[string]*base.Node
	routines       map[base.RoutineId]*base.Routine
	connections    map[string]*net.TCPConn
	listener       *net.TCPListener
	routineCounter *uint64
}

// 构建 godist.Agent 对象，返回其指针。
func New(node string) *Agent {
	nameAndHost := make([]string, 2)
	nameAndHost = strings.SplitN(node, "@", 2)
	gpmd := base.GPMD{
		Host: "",
		Port: EPMD_PORT,
	}
	return &Agent{
		gpmd:           gpmd,
		name:           nameAndHost[0],
		host:           nameAndHost[1],
		nodes:          make(map[string]*base.Node),
		routines:       make(map[base.RoutineId]*base.Routine),
		connections:    make(map[string]*net.TCPConn),
		routineCounter: &routineCounter,
	}
}

func (a *Agent) Host() string {
	return a.host
}

func (a *Agent) Port() uint16 {
	return a.port
}

func (a *Agent) Name() string {
	return a.name
}

func (a *Agent) Node() *base.Node {
	return &base.Node{
		Port: a.Port(),
		Name: a.Name(),
		Host: a.Host(),
	}
}

// 设置本机的 GPMD 服务地址。默认为 ":2613"
func (a *Agent) SetGPMD(host string, port uint16) {
	a.gpmd.Host = host
	a.gpmd.Port = port
}

// 向 agent 注册一个 Goroutine 。如果该 Goroutine 对象已经被设置过 Id ，则会抛出
// panic 。
func (agent *Agent) RegisterRoutine(routine *base.Routine) {
	routine.SetId(agent.incrRoutineId())
	agent.registerRoutine(routine)
}

func (agent *Agent) Stop() {
	agent.listener.Close()
	agent.Unregister()
}

func (agent *Agent) Unregister() {
	resolvedAddr, rErr := net.ResolveTCPAddr("tcp", agent.gpmd.Address())
	if rErr != nil {
		panic(fmt.Sprintf("godist: GPMD address error: %s", rErr))
	}
	conn, dErr := net.DialTCP("tcp", nil, resolvedAddr)
	if dErr != nil {
		panic(fmt.Sprintf("godist: GPMD dial error: %s", dErr))
	}
	requestBuf := new(bytes.Buffer)
	binpacker.NewPacker(requestBuf).
		PushByte(gpmd.REQ_UNREGISTER).
		PushUint16(uint16(len(agent.name))).
		PushString(agent.name)
	request := binpacker.AddUint16Perfix(requestBuf.Bytes())
	if _, wErr := conn.Write(request); wErr != nil {
		panic(fmt.Sprintf("godist: Send unregister message error: %s", wErr))
	}
	var apiCode, resCode byte
	unpacker := binpacker.NewUnpacker(conn)
	unpacker.FetchByte(&apiCode).FetchByte(&resCode)
	if unpacker.Error() != nil || apiCode != gpmd.REQ_UNREGISTER || resCode != gpmd.ACK_RES_OK {
		log.Printf("godist: unregister node %s@%s error", agent.name, agent.host)
	}
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
	requestBuf := new(bytes.Buffer)
	binpacker.NewPacker(requestBuf).
		PushByte(gpmd.REQ_REGISTER).
		PushUint16(agent.port).
		PushUint16(uint16(len(agent.name))).
		PushString(agent.name).
		PushUint16(uint16(len(agent.host))).
		PushString(agent.host)
	request := binpacker.AddUint16Perfix(requestBuf.Bytes())
	if _, wErr := conn.Write(request); wErr != nil {
		panic(fmt.Sprintf("godist: Send register message error: %s", wErr))
	}
	unpacker := binpacker.NewUnpacker(conn)
	var apiCode, resCode byte
	unpacker.FetchByte(&apiCode).FetchByte(&resCode)
	if unpacker.Error() != nil || apiCode != gpmd.REQ_REGISTER || resCode != gpmd.ACK_RES_OK {
		panic(fmt.Sprintf(
			"godist: Register failed. API: %d, Res: %d",
			apiCode,
			resCode,
		))
	}
}

func (agent *Agent) QueryAllNode(nodeName string) {
	name, _ := parseNameAndHost(nodeName)
	if name == agent.name {
		return
	}
	if conn, exist := agent.connections[name]; exist {
		requestBuf := new(bytes.Buffer)
		binpacker.NewPacker(requestBuf).
			PushByte(REQ_QUERY_ALL).
			PushUint16(uint16(len(agent.name))).
			PushString(agent.name)
		request := binpacker.AddUint64Perfix(requestBuf.Bytes())
		if _, err := conn.Write(request); err != nil {
			return
		}
		// ANSWER
		unpacker := binpacker.NewUnpacker(conn)
		var ackCode byte
		unpacker.FetchByte(&ackCode)
		if unpacker.Error() != nil || ackCode != ACK_QUERY_ALL_OK {
			return
		}
		count, err := unpacker.ShiftUint16()
		if err != nil {
			return
		}
		for i := 0; i < int(count); i++ {
			var port uint16
			var name, host string
			unpacker.FetchUint16(&port).
				StringWithUint16Perfix(&name).
				StringWithUint16Perfix(&host)
			if unpacker.Error() != nil {
				return
			}
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

// 向目标节点的 GPMD 查询节点的端口号等详细信息。
//  `nodeName` e.g. "player_01@player.1.example.local"
func (agent *Agent) QueryNode(nodeName string) {
	name, host := parseNameAndHost(nodeName)
	if name == agent.name {
		return
	}
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
		requestBuf := new(bytes.Buffer)
		binpacker.NewPacker(requestBuf).
			PushByte(gpmd.REQ_QUERY).
			PushUint16(uint16(len(name))).
			PushString(name)
		request := binpacker.AddUint16Perfix(requestBuf.Bytes())
		if _, wErr := conn.Write(request); wErr != nil {
			return
		}
		unpacker := binpacker.NewUnpacker(conn)
		var ackCode, resCode byte
		if unpacker.FetchByte(&ackCode).FetchByte(&resCode).Error() != nil {
			return
		}
		if ackCode != gpmd.REQ_QUERY || resCode != gpmd.ACK_RES_OK {
			return
		}
		var port uint16
		var ackName string
		if unpacker.
			FetchUint16(&port).
			StringWithUint16Perfix(&ackName).
			Error() != nil {
			return
		}
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

// XXX: 权衡参数传入格式是否需要是节点全名(xx@xx)还是节点名(xx)即可。
//
// 尝试向目标节点建立连接。该节点名称必须在 `agent.nodes` 中有注册的信息。建立好
// 之后会一直保持持有连接。用于向目标节点的 Goroutine 消息发送。
//  `nodeName` e.g. "player_01@player.1.example.local"
func (agent *Agent) ConnectTo(nodeName string) {
	agent.connectTo(nodeName, false)
}

func (agent *Agent) connectTo(nodeName string, isReturn bool) {
	name, _ := parseNameAndHost(nodeName)
	if name == agent.name {
		return
	}
	if node, exist := agent.nodes[name]; exist {
		address, rErr := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", node.Host, node.Port))
		if rErr != nil {
			// handle error
			return
		}
		conn, dErr := net.DialTCP("tcp", nil, address)
		if dErr != nil {
			// handle error
			return
		}
		requestBuf := new(bytes.Buffer)
		pk := binpacker.NewPacker(requestBuf).PushByte(REQ_CONN)
		if isReturn {
			pk.PushByte(ACK_CONN_IS_RETURN)
		} else {
			pk.PushByte(ACK_CONN_IS_NOT_RETURN)
		}
		pk.PushUint16(agent.port).
			PushUint16(uint16(len(agent.name))).
			PushString(agent.name).
			PushUint16(uint16(len(agent.host))).
			PushString(agent.host)
		request := binpacker.AddUint64Perfix(requestBuf.Bytes())
		if _, wErr := conn.Write(request); wErr != nil {
			// handle error
			conn.Close()
			return
		}
		unpacker := binpacker.NewUnpacker(conn)
		ackCode, err := unpacker.ShiftByte()
		if err != nil || ackCode != ACK_CONN_OK {
			conn.Close()
			return
		}
		agent.registerConn(name, conn)
	}
}

// 向目标 Goroutine 发送消息。该目标节点连接必须事先注册在 `agent.connections`
// 中。
func (agent *Agent) CastTo(nodeName string, routineId base.RoutineId, message []byte) {
	log.Printf("godist: Cast to %d@%s message...", uint64(routineId), nodeName)
	if conn, exist := agent.connections[nodeName]; exist {
		requestBuf := new(bytes.Buffer)
		binpacker.NewPacker(requestBuf).
			PushByte(REQ_CAST).
			PushUint64(uint64(routineId)).
			PushUint64(uint64(len(message))).
			PushBytes(message)
		request := binpacker.AddUint64Perfix(requestBuf.Bytes())
		if _, wErr := conn.Write(request); wErr != nil {
			return
		}
		unpacker := binpacker.NewUnpacker(conn)
		ackCode, err := unpacker.ShiftByte()
		if err != nil || ackCode != ACK_CAST_OK {
			return
		}
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
		log.Printf("godist: Node %s register...", node.Name)
	}
}

func (agent *Agent) registerConn(name string, conn *net.TCPConn) {
	oldConn, exist := agent.connections[name]
	if exist {
		log.Printf("godist: Close the old connection of node %s", name)
		oldConn.Close()
	}
	agent.connections[name] = conn
	log.Printf("godist: Hoding node %s connection", name)
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
