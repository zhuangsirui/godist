package godist

import (
	"godist/base"
	"godist/gpmd"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var tAgent1, tAgent2, tAgent3 *Agent

var nodeName1, nodeName2, nodeName3 = "testnode1@localhost", "testnode2@localhost", "testnode3@localhost"

func TestAgent(t *testing.T) {
	convey.Convey("Agent", t, func() {

		convey.Convey("Error host", func() {
			agent := New("xx@x8x8x*(&)")
			agent.SetGPMD("fake**xx", 1000)
			convey.So(func() {
				agent.Register()
			}, convey.ShouldPanic)
			convey.So(func() {
				agent.Listen()
			}, convey.ShouldPanic)
		})

		var gpmdPort uint16 = 1988
		agent := New("agent@localhost")
		convey.So(agent.Host(), convey.ShouldEqual, agent.Host())
		convey.So(agent.Name(), convey.ShouldEqual, agent.Name())
		agent.SetGPMD("localhost", gpmdPort)

		convey.Convey("Register panic", func() {
			convey.So(func() {
				agent.Register()
			}, convey.ShouldPanic)
		})

		convey.Convey("Start GPMD", func() {
			m := gpmd.New(agent.Host(), gpmdPort)
			m.Serve()

			convey.Convey("Register", func() {

				convey.So(func() {
					agent.Listen()
					agent.Register()
					go agent.Serve()
				}, convey.ShouldNotPanic)

				convey.Convey("Connect", func() {
					targetAgent := New("target_agent@localhost")
					targetAgent.SetGPMD(agent.Host(), gpmdPort)
					convey.So(func() {
						targetAgent.Listen()
						targetAgent.Register()
						go targetAgent.Serve()
					}, convey.ShouldNotPanic)

					agent.QueryNode(targetAgent.node.FullName())
					agent.ConnectTo(targetAgent.Name())
					convey.So(agent.nodeExist(targetAgent.Name()), convey.ShouldBeTrue)
					convey.So(agent.connections, convey.ShouldContainKey, targetAgent.Name())

					targetAgent.QueryNode(agent.node.FullName())
					targetAgent.ConnectTo(agent.Name())
					convey.So(targetAgent.nodeExist(agent.Name()), convey.ShouldBeTrue)
					convey.So(targetAgent.connections, convey.ShouldContainKey, agent.Name())

					convey.Convey("Register routine", func() {
						routine := &base.Routine{
							Channel: make(chan []byte),
						}
						agent.RegisterRoutine(routine)
						routine2, exist := agent.findRoutine(routine.GetId())
						convey.So(exist, convey.ShouldBeTrue)
						convey.So(routine2, convey.ShouldEqual, routine)
					})

					convey.Convey("Cast to", func() {
						routine := &base.Routine{
							Channel: make(chan []byte, 1),
						}
						ping := []byte("ping")
						targetAgent.RegisterRoutine(routine)
						go agent.CastTo(targetAgent.Name(), routine.GetId(), ping)
						convey.So(<-routine.Channel, convey.ShouldResemble, ping)
					})

					convey.Convey("Query all", func() {
						another := New("another@localhost")
						another.SetGPMD(another.Host(), gpmdPort)
						another.Listen()
						another.Register()
						go another.Serve()
						another.QueryNode(agent.Name())
						another.ConnectTo(agent.Name())
						another.QueryAllNode(agent.Name())
						convey.So(another.nodeExist(targetAgent.Name()), convey.ShouldBeTrue)
					})
				})

			})

			agent.Stop()
			m.Stop()
			m.Stopped()
		})
	})
}

/*
func testQueryAll(t *testing.T) {
	tAgent3 = New(nodeName3)
	tAgent3.Listen()
	go tAgent3.Serve()
	tAgent3.QueryNode(nodeName2)
	tAgent3.ConnectTo(nodeName2)
	tAgent3.QueryAllNode(nodeName2)
	name, _ := parseNameAndHost(nodeName1)
	if !tAgent3.nodeExist(name) {
		t.Error("Node 1 dosen't connect automatic")
	}
}

func testCastTo(t *testing.T) {
	c := make(chan []byte) // make channle sync for test
	routine := &base.Routine{
		Channel: c,
	}
	tAgent1.RegisterRoutine(routine)
	name, _ := parseNameAndHost(nodeName1)
	go tAgent2.CastTo(name, routine.GetId(), []byte{'p', 'i', 'n', 'g'})
	<-c
}

func testRegisterRoutine(t *testing.T) {
	c := make(chan []byte)
	routine := &base.Routine{
		Channel: c,
	}
	tAgent1.RegisterRoutine(routine)
	routine2, exist := tAgent1.findRoutine(routine.GetId())
	if !exist {
		t.Error("register routine failed.")
	}
	if routine != routine2 {
		t.Error("routine is diffrent!")
	}
}

func testNew(t *testing.T) {
	tAgent1 = New(nodeName1)
	tAgent1.Host()
}

func testSetGPMD(t *testing.T) {
	tAgent1.SetGPMD("localhost", 2613)
}

func testRegisterToGPMD(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Now you should panic when there's no panic!")
		}
	}()
	tAgent1.Register()
}

func testRegisterToGPMD2(t *testing.T) {
	gpmd.Init()
	tAgent1.Listen()
	tAgent1.Register()
}

func testConnect(t *testing.T) {
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
	if _, exist := tAgent1.findConn(name2); !exist {
		t.Error("connection not exist")
	}
	if !tAgent2.nodeExist(name1) {
		t.Error("node not exist")
	}
	if _, exist := tAgent2.findConn(name1); !exist {
		t.Error("connection not exist")
	}
}

//func testRegisterRoutine(t *testing.T) {
//c := make(chan []byte)
//routine := &base.Routine{
//Channel: c,
//}
//tAgent1.RegisterRoutine(routine)
//routine2, exist := tAgent1.findRoutine(routine.GetId())
//if !exist {
//t.Error("register routine failed.")
//}
//if routine != routine2 {
//t.Error("routine is diffrent!")
//}
//}

func testStop(t *testing.T) {
	_a := New("testAgentForStop@localhost")
	_a.Listen()
	go _a.Serve()
	_a.Stop()
}
*/
