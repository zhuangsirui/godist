package gpmd

import (
	"fmt"
	"godist/base"
	"net"
	"orbit/log"
	"sync"
)

// GPMD Manager 对象。保持本机的所有节点，以及向外部节点提供查询服务。
type Manager struct {
	port     uint16
	host     string
	nodes    map[string]*base.Node
	nodeLock *sync.RWMutex
	listener *net.TCPListener
}

// 创建一个 GPMD 实例。
func New(host string, port uint16) *Manager {
	return &Manager{
		port:     port,
		host:     host,
		nodes:    make(map[string]*base.Node),
		nodeLock: new(sync.RWMutex),
	}
}

// 启动 GPMD 服务。
func (m *Manager) Serve() {
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
	m.listener = listener
	log.Debugf("GPMD started on %s", m.listener.Addr())
	go m.acceptLoop()
}

func (m *Manager) Stop() {
	m.listener.Close()
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
		log.Infof("Node %s register success", node.Name)
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
