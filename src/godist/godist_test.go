package godist

import(
	"testing"
	"godist/gpmd"
)

var tAgent1, tAgent2 *Agent

var nodeName1, nodeName2 = "testnode1@localhost", "testnode2@localhost"

func TestNew(t *testing.T) {
	tAgent1 = New(nodeName1)
}

func TestSetGPMD(t *testing.T) {
	tAgent1.SetGPMD("localhost", 2613)
}

func TestRegisterToGPMD(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Now you should panic when there's no panic!")
		}
	}()
	tAgent1.Register()
}

func TestRegisterToGPMD2(t *testing.T) {
	gpmd.Init()
	tAgent1.Listen()
	tAgent1.Register()
}

func TestConnect(t *testing.T) {
	tAgent2 = New(nodeName2)
	tAgent2.Listen()
	go tAgent1.Serve()
	go tAgent2.Serve()
	tAgent2.Register()
	tAgent1.ConnectTo(nodeName2)
	t.Log(tAgent1.nodes)
	t.Log(tAgent1.connections)
}
