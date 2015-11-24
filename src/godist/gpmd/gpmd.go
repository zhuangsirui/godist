package gpmd

import(
	"fmt"
	"net"
	"godist/base"
)

type manager struct {
	port uint16
	host string
	nodes map[string]*base.Node
}

var m = &manager{
	port: 2613,
	host: "",
	nodes: make(map[string]*base.Node),
}

func SetPort(port uint16) {
	m.port = port
}

func SetHost(host string) {
	m.host = host
}

func find(name string) (*base.Node, bool) {
	node, exist := m.nodes[name]
	return node, exist
}

func register(node *base.Node) bool {
	_, exist := m.nodes[node.Name]
	if !exist {
		m.nodes[node.Name] = node
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
