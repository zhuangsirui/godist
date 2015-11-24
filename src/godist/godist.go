package godist

import(
	"net"
	"godist/base"
)

type Agent struct {
	name string
	host string
	port uint16
	lisener *net.TCPListener
	nodes map[string]*base.Node
	routines map[base.RoutineId]*base.Routine
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
