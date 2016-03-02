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

				convey.Convey("Cast Local", func() {
					ping := []byte("ping")
					CastTo("static", routine.GetId(), ping)
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
		})

		m.Stop()
		m.Stopped()
	})
}

/*
var nodeName4, nodeName5 = "testnode4@localhost", "testnode5@localhost"
var tAgent5 = New(nodeName5)

func TestStaticSetGPMD(t *testing.T) {
	SetGPMD("localhost", 2613)
	Host()
	Name()
	Port()
	Node()
	NewProcess()
}

func TestStaticInit(t *testing.T) {
	Init(nodeName4)
}

func TestStaticConnectTo(t *testing.T) {
	tAgent5.Listen()
	go tAgent5.Serve()
	tAgent5.Register()
	ConnectTo(nodeName5)
	QueryAllNode(nodeName5)
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

func TestStaticStop(t *testing.T) {
	Stop()
}

func TestCastToLocal(t *testing.T) {
	p := NewProcess()
	go CastToLocal(p.routine.GetId(), []byte{'p', 'i', 'n', 'g'})
	<-p.Channel
}
*/
