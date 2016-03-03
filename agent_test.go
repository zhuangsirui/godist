package godist

import (
	"bytes"
	"fmt"
	"godist/base"
	"godist/gpmd"
	"io/ioutil"
	"log"
	"net"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/zhuangsirui/binpacker"
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

				convey.Convey("Query self", func() {
					agent.QueryAllNode(agent.node.FullName())
				})

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

					convey.Convey("Cast to", func() {
						routine := &base.Routine{
							Channel: make(chan []byte, 1),
						}
						ping := []byte("ping")
						targetAgent.RegisterRoutine(routine)
						go agent.CastTo(targetAgent.Name(), routine.GetId(), ping)
						convey.So(<-routine.Channel, convey.ShouldResemble, ping)
						go agent.CastTo(targetAgent.Name(), base.RoutineId(9898), ping)
					})

					targetAgent.Stop()
					targetAgent.Stopped()
				})

				convey.Convey("Bad request", func() {
					requestBuf := new(bytes.Buffer)
					conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", agent.Host(), agent.Port()))
					convey.So(err, convey.ShouldBeNil)
					binpacker.NewPacker(requestBuf).PushByte(0xff)
					_, err = conn.Write(binpacker.AddUint64Perfix(requestBuf.Bytes()))
					convey.So(err, convey.ShouldBeNil)
				})

				convey.Convey("Bad connection", func() {
					requestBuf := new(bytes.Buffer)
					conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", agent.Host(), agent.Port()))
					convey.So(err, convey.ShouldBeNil)
					binpacker.NewPacker(requestBuf).PushByte(0xff)
					_, err = conn.Write(binpacker.AddUint64Perfix(requestBuf.Bytes()))
					conn.Close()
					convey.So(err, convey.ShouldBeNil)
				})

				agent.Stop()
				agent.Stopped()
			})

			m.Stop()
			m.Stopped()
		})

		convey.Convey("Unregister", func() {
			localhost := "localhost"
			var gpmdPort uint16 = 1989
			m := gpmd.New(localhost, gpmdPort)
			m.Serve()
			agent := New("agent@localhost")
			agent.SetGPMD("localhost", gpmdPort)
			agent.Register()

			convey.Convey("GPMD Addr error", func() {
				agent.SetGPMD("fake**host", 0)
				convey.So(func() {
					agent.Unregister()
				}, convey.ShouldNotPanic)
			})

			m.Stop()
			m.Stopped()

			convey.Convey("GPMD down", func() {
				convey.So(func() {
					agent.Unregister()
				}, convey.ShouldNotPanic)
			})
		})

	})
}
