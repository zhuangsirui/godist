package godist

import(
	"testing"
	"godist/base"
)

var nodeName4, nodeName5 = "testnode4@localhost", "testnode5@localhost"
var tAgent5 = New(nodeName5)

func TestStaticSetGPMD(t *testing.T) {
	SetGPMD("localhost", 2613)
}

func TestStaticInit(t *testing.T) {
	Init(nodeName4)
}

func TestStaticRegisterToGPMD(t *testing.T) {
	Register()
}

func TestStaticConnectTo(t *testing.T) {
	tAgent5.Listen()
	go tAgent5.Serve()
	tAgent5.Register()
	ConnectTo(nodeName5)
}

func TestStaticCastTo(t *testing.T) {
	tAgent5.QueryNode(nodeName4)
	tAgent5.ConnectTo(nodeName4)
	c := make(chan []byte) // make channle sync for test
	routine := &base.Routine{
		Channel: c,
	}
	tAgent5.RegisterRoutine(routine)
	name, _ := parseNameAndHost(nodeName5)
	go CastTo(name, routine.GetId(), []byte{'p', 'i', 'n', 'g'})
	<-c
}

