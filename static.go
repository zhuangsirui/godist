package godist

import (
	"godist/base"
)

var _agent = &Agent{}

// 在本进程运行一个 agent 服务。该服务保存了自身进程中向他注册的所有 Goroutine
// 的消息通道。
//
// 该服务会监听一个随机端口，并且向本机的 GPMD 发送一条注册消息，将监听的随机
// 端口注册。
//
// 当需要向集群上其他 Goroutine 发送消息时，需要知道该 Goroutine 的宿主节点名称
// 以及该 Goroutine 的 ID 。然后调用 `godist.Cast(hostname, routineId, message)`
// 向目标 Goroutine 发送消息。消息格式是 []byte 。
func Init(name string) {
	_agent = New(name)
	_agent.Listen()
	go _agent.Serve()
}

// 设置本机的 GPMD 服务地址。默认为 ":2613"
func SetGPMD(host string, port uint16) {
	_agent.gpmd.Host = host
	_agent.gpmd.Port = port
}

func Host() string {
	return _agent.host
}

// 向本地 GPMD 注册节点信息，无法注册会 panic 异常。
func Register() {
	_agent.Register()
}

// 向本地的 agent 注册一个 Goroutine 。如果该 Goroutine 对象已经被设置过 Id ，则
// 会抛出 panic 。
func RegisterRoutine(routine *base.Routine) {
	routine.SetId(_agent.incrRoutineId())
	_agent.registerRoutine(routine)
}

// 尝试向另一个节点建立连接。建立好之后会一直保持连接。用于节点之间的 Goroutine
// 消息收发。
func ConnectTo(nodeName string) {
	_agent.QueryNode(nodeName)
	_agent.ConnectTo(nodeName)
}

// 向目标 Goroutine 发送消息。
func CastTo(nodeName string, routineId base.RoutineId, message []byte) {
	_agent.CastTo(nodeName, routineId, message)
}

func CastToLocal(routineId base.RoutineId, message []byte) {
	_agent.CastTo(_agent.name, routineId, message)
}
