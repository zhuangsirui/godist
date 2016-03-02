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
