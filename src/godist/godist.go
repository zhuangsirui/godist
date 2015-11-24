package godist

import(
	"net"
)

type Agent struct {
	name string
	host string
	port uint16
	lisener *net.TCPListener
	routines map[routineId]*routine
}

var agent = &Agent{
	routines: make(map[routineId]*routine),
}

func find(routineId routineId) (*routine, bool) {
	routine, exist := agent.routines[routineId]
	return routine, exist
}
