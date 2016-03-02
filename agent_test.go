package godist

import (
	"godist/base"
	"godist/gpmd"
	"io/ioutil"
	"log"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func testMultiTimes(t *testing.T) {
	for i := 0; i < 50; i++ {
		TestAgent(t)
		TestNewProcess(t)
		TestStatic(t)
	}
}

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

		convey.Convey("Without GPMD", func() {
			agent := New("agent@localhost")
			convey.So(agent.Host(), convey.ShouldEqual, agent.Host())
			convey.So(agent.Name(), convey.ShouldEqual, agent.Name())
			agent.SetGPMD("localhost", 1000)
			convey.So(func() {
				agent.Register()
			}, convey.ShouldPanic)
		})

		convey.Convey("With GPMD", func() {
			localhost := "localhost"
			var gpmdPort uint16 = 1989
			m := gpmd.New(localhost, gpmdPort)
			m.Serve()

			convey.Convey("Start base agent", func() {
				agent := New("agent@localhost")
				convey.So(agent.Host(), convey.ShouldEqual, agent.Host())
				convey.So(agent.Name(), convey.ShouldEqual, agent.Name())
				agent.SetGPMD("localhost", gpmdPort)
				convey.So(func() {
					agent.Listen()
					agent.Register()
				}, convey.ShouldNotPanic)
				go agent.Serve()

				convey.Convey("Connect", func() {
					targetAgent := New("target_agent@localhost")
					targetAgent.SetGPMD(agent.Host(), gpmdPort)
					targetAgent.Listen()
					targetAgent.Register()
					go targetAgent.Serve()

					agent.QueryNode(targetAgent.node.FullName())
					agent.ConnectTo(targetAgent.Name())
					convey.So(agent.nodeExist(targetAgent.Name()), convey.ShouldBeTrue)
					convey.So(agent.connections, convey.ShouldContainKey, targetAgent.Name())

					targetAgent.QueryNode(agent.node.FullName())
					targetAgent.ConnectTo(agent.Name())
					convey.So(targetAgent.nodeExist(agent.Name()), convey.ShouldBeTrue)
					convey.So(targetAgent.connections, convey.ShouldContainKey, agent.Name())

					convey.Convey("Cast to", func() {
						routine := &base.Routine{
							Channel: make(chan []byte, 1),
						}
						ping := []byte("ping")
						targetAgent.RegisterRoutine(routine)
						go agent.CastTo(targetAgent.Name(), routine.GetId(), ping)
						convey.So(<-routine.Channel, convey.ShouldResemble, ping)

						convey.Convey("Query all", func() {
							another := New("another@localhost")
							another.SetGPMD(localhost, gpmdPort)
							another.Listen()
							another.Register()
							go another.Serve()
							another.QueryNode(agent.Name())
							another.ConnectTo(agent.Name())
							another.QueryAllNode(agent.Name())
							convey.So(another.nodeExist(targetAgent.Name()), convey.ShouldBeTrue)
							another.Stop()
							another.Stopped()
						})
					})

					targetAgent.Stop()
					targetAgent.Stopped()
				})

				agent.Stop()
				agent.Stopped()
			})

			m.Stop()
			m.Stopped()
		})

	})
}
