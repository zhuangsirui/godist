package gpmd

import (
	"fmt"
	"godist/base"
	"log"
	"net"
)

type manager struct {
	port  uint16
	host  string
	nodes map[string]*base.Node
}

var m = &manager{
	port:  2613,
	host:  "",
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
	log.Printf("GPMD started on %s", listener.Addr())
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
		log.Printf("Node %s register success", node.Name)
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
		conn, err := listener.AcceptTCP()
		if err != nil {
			// TODO handle accept error
			log.Printf("Accept error %s", err)
			break
		}
		// 同步调用，原子性处理各个节点的请求
		log.Printf("Handle connection...")
		handleConnection(conn)
		conn.Close()
	}
}
