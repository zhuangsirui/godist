package gpmd

import (
	"fmt"
	"godist/base"
	"log"
	"net"
	"sync"
)

// GPMD Manager 对象。保持本机的所有节点，以及向外部节点提供查询服务。
type Manager struct {
	port      uint16
	host      string
	nodes     map[string]*base.Node
	nodeLock  *sync.RWMutex
	listener  *net.TCPListener
	isStop    bool
	stopped   chan bool
	restarted chan bool
}

// 创建一个 GPMD 实例。
func New(host string, port uint16) *Manager {
	return &Manager{
		port:      port,
		host:      host,
		nodes:     make(map[string]*base.Node),
		nodeLock:  new(sync.RWMutex),
		stopped:   make(chan bool, 1),
		restarted: make(chan bool, 1),
	}
}

func (m *Manager) Stopped() {
	<-m.stopped
}

func (m *Manager) Restarted() {
	<-m.restarted
}

// 启动 GPMD 服务。
func (m *Manager) Serve() {
	listenAddr, rErr := net.ResolveTCPAddr(
		"tcp",
		fmt.Sprintf("%s:%d", m.host, m.port),
	)
	if rErr != nil {
		log.Panicf("GPMD resolve tcp address error: %s", rErr)
	}
	listener, listenErr := net.ListenTCP("tcp", listenAddr)
	if listenErr != nil {
		log.Panicf("GPMD listen port error: %s", listenErr)
	}
	m.listener = listener
	log.Printf("GPMD started on %s", m.listener.Addr())
	go m.acceptLoop()
}

func (m *Manager) find(name string) (*base.Node, bool) {
	m.nodeLock.RLock()
	defer m.nodeLock.RUnlock()
	node, exist := m.nodes[name]
	return node, exist
}

func (m *Manager) register(node *base.Node) bool {
	m.nodeLock.Lock()
	defer m.nodeLock.Unlock()
	_, exist := m.nodes[node.Name]
	if !exist {
		m.nodes[node.Name] = node
		log.Printf("GPMD node %s register success", node.Name)
	}
	return !exist
}

func (m *Manager) unregister(name string) bool {
	m.nodeLock.Lock()
	defer m.nodeLock.Unlock()
	_, exist := m.nodes[name]
	if exist {
		delete(m.nodes, name)
	}
	return exist
}
