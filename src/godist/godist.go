package godist

import(
	"fmt"
	"net"
)

type manager struct {
	port int
	host string
	nodes map[string]node
}

var m = &manager{
	port: 65432,
	host: "127.0.0.1",
	nodes: make(map[string]node),
}

func SetPort(port int) {
	m.port = port
}

func SetHost(host string) {
	m.host = host
}

func find(name string) (node, bool) {
	node, exist := m.nodes[name]
	return node, exist
}

func register(node node) bool {
	_, exist := m.nodes[node.name]
	if !exist {
		m.nodes[node.name] = node
	}
	return !exist
}

func unregister(name string) bool {
	_, exist := m.nodes[name]
	if exist {
		delete(m.nodes, name)
	}
	return exist
}

func Init() {
	listenAddr, rErr := net.ResolveTCPAddr(
		"tcp",
		fmt.Sprintf("%s:%d", m.host, m.port),
	)
	if rErr != nil {
		panic(fmt.Sprintf("GPMD listen port error: ", rErr))
	}
	listener, listenErr := net.ListenTCP("tcp", listenAddr)
	if listenErr != nil {
		panicInfo := fmt.Sprintf("GPMD listen port error: ", listenErr)
		panic(panicInfo)
	}
	for {
		conn, acceptErr := listener.AcceptTCP()
		if acceptErr != nil {
			// TODO handle accept error
			break
		}
		defer conn.Close()
		// 同步调用，原子性处理各个节点的请求
		handleRequest(conn)
	}
}
