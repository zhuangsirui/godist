package godist

import(
	"testing"
	"godist/base"
)

var nodeName3, nodeName4 = "testnode3@localhost", "testnode4@localhost"
var tAgent4 = New(nodeName4)

func TestStaticSetGPMD(t *testing.T) {
	SetGPMD("localhost", 2613)
}

func TestStaticInit(t *testing.T) {
	Init(nodeName3)
}

func TestStaticRegisterToGPMD(t *testing.T) {
	Register()
}

func TestStaticConnectTo(t *testing.T) {
	tAgent4.Listen()
	go tAgent4.Serve()
	tAgent4.Register()
	ConnectTo(nodeName4)
}

func TestStaticCastTo(t *testing.T) {
	tAgent4.QueryNode(nodeName3)
	tAgent4.ConnectTo(nodeName3)
	c := make(chan []byte) // make channle sync for test
	routine := &base.Routine{
		Channel: c,
	}
	tAgent4.RegisterRoutine(routine)
	name, _ := parseNameAndHost(nodeName4)
	go CastTo(name, routine.GetId(), []byte{'p', 'i', 'n', 'g'})
	<-c
}

