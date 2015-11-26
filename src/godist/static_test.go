package godist

import(
	"testing"
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
	ConnectTo(nodeName4)
}

