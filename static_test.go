package godist

import (
	"errors"
	"godist/base"
	"godist/gpmd"
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestStatic(t *testing.T) {
	convey.Convey("Static", t, func() {
		testhost := "localhost"
		var gpmdPort uint16 = 1777
		m := gpmd.New(testhost, gpmdPort)
		m.Serve()

		convey.Convey("Init", func() {
			Init("static@localhost")
			SetGPMD(testhost, gpmdPort)
			Register()
			convey.So(Host(), convey.ShouldEqual, testhost)
			convey.So(Name(), convey.ShouldEqual, "static")
			convey.So(Port(), convey.ShouldEqual, _agent.Port())
			convey.So(Node(), convey.ShouldResemble, _agent.Node())

			convey.Convey("Register routine", func() {
				routine := &base.Routine{
					Channel: make(chan []byte, 1),
				}
				RegisterRoutine(routine)

				convey.Convey("Cast local", func() {
					ping := []byte("ping")
					CastTo(_agent.Name(), routine.GetId(), ping)
					convey.So(<-routine.Channel, convey.ShouldResemble, ping)
				})

				convey.Convey("Connect", func() {
					agent := New("static2@localhost")
					agent.Listen()
					agent.SetGPMD(agent.Host(), gpmdPort)
					agent.Register()
					go agent.Serve()
					ConnectTo(agent.Node().FullName())
					QueryAllNode(agent.Node().FullName())

					convey.Convey("Cast remote", func() {
						routine := &base.Routine{
							Channel: make(chan []byte, 1),
						}
						ping := []byte("ping")
						agent.RegisterRoutine(routine)
						CastTo(agent.Name(), routine.GetId(), ping)
						convey.So(<-routine.Channel, convey.ShouldResemble, ping)
					})
					agent.Stop()
					agent.Stopped()
				})
			})

			convey.Convey("Static process", func() {
				process := NewProcess()
				ping := []byte("ping")
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					process.Run(func(message []byte) error {
						convey.Convey("in static process", t, func() {
							convey.So(message, convey.ShouldResemble, ping)
						})
						return errors.New("stop")
					})
					wg.Done()
				}()
				process.Channel <- ping
				wg.Wait()
			})

			Stop()
			Stopped()
		})

		m.Stop()
		m.Stopped()
	})
}
