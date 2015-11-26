package godist

import(
	"testing"
	"godist/gpmd"
	"godist/base"
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
	tAgent1.QueryNode(nodeName2)
	tAgent1.ConnectTo(nodeName2)
	tAgent2.QueryNode(nodeName1)
	tAgent2.ConnectTo(nodeName1)
	name1, _ := parseNameAndHost(nodeName1)
	name2, _ := parseNameAndHost(nodeName2)
	if !tAgent1.nodeExist(name2) {
		t.Error("node not exist")
	}
	if _, exist := tAgent1.connections[name2]; !exist {
		t.Error("connection not exist")
	}
	if !tAgent2.nodeExist(name1) {
		t.Error("node not exist")
	}
	if _, exist := tAgent2.connections[name1]; !exist {
		t.Error("connection not exist")
	}
}

func TestRegisterRoutine(t *testing.T) {
	c := make(chan []byte)
	routine := &base.Routine{
		Channel: c,
	}
	tAgent1.RegisterRoutine(routine)
	routine2, exist := tAgent1.routines[routine.GetId()]
	if !exist {
		t.Error("register routine failed.")
	}
	if routine != routine2 {
		t.Error("routine is diffrent!")
	}
}

func TestCastTo(t *testing.T) {
	c := make(chan []byte) // make channle sync for test
	routine := &base.Routine{
		Channel: c,
	}
	tAgent1.RegisterRoutine(routine)
	name, _ := parseNameAndHost(nodeName1)
	go tAgent2.CastTo(name, routine.GetId(), []byte{'p', 'i', 'n', 'g'})
	<-c
}
