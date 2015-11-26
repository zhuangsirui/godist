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

// 初始化 GPMD 服务。
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
	go acceptLoop(listener)
}

// 设置 GPMD 的监听端口。默认 2613 。
func SetPort(port uint16) {
	m.port = port
}

// 设置 GPMD 的监听地址。默认为空。
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

func acceptLoop(listener *net.TCPListener) {
	for {
		conn, acceptErr := listener.AcceptTCP()
		if acceptErr != nil {
			// TODO handle accept error
			break
		}
		defer conn.Close()
		// 同步调用，原子性处理各个节点的请求
		handleConnection(conn)
	}
}
