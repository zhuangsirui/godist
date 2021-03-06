package godist

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewProcess(t *testing.T) {
	convey.Convey("Process", t, func() {
		convey.Convey("Process common test", func() {
			node := "process@localhost"
			agent := New(node)
			process := agent.NewProcess()
			convey.So(process.GetId(), convey.ShouldEqual, process.routine.GetId())
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				process.Run(func(message []byte) error {
					// convey needs context in new goroutine, so I need to do this.
					convey.Convey("in process", t, func() {
						convey.So(message, convey.ShouldResemble, []byte("ping"))
					})
					return errors.New("stop")
				})
				wg.Done()
			}()
			process.Channel <- []byte("ping")
			wg.Wait()
		})

		convey.Convey("Process restart", func() {
			node := "process_2@localhost"
			agent := New(node)
			process := agent.NewProcess()
			convey.So(process.GetId(), convey.ShouldEqual, process.routine.GetId())
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				process.Run(func(message []byte) error {
					if bytes.Compare(message, []byte("panic")) == 0 {
						panic("just a panic")
					} else {
						wg.Done()
						return errors.New("stop")
					}
				})
			}()
			process.Channel <- []byte("panic")
			process.Channel <- []byte("peace")
			wg.Wait()
		})
	})
}
