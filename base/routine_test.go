package base

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRoutine(t *testing.T) {
	convey.Convey("Init Routine", t, func() {
		c := make(chan []byte)
		r := Routine{
			Channel: c,
		}
		convey.So(func() {
			r.GetId()
		}, convey.ShouldPanic)
		id := RoutineId(3312)
		convey.So(func() {
			r.SetId(id)
		}, convey.ShouldNotPanic)
		convey.So(func() {
			r.SetId(id)
		}, convey.ShouldPanic)
		convey.So(r.GetId(), convey.ShouldEqual, id)
		go r.Cast([]byte{'p', 'i', 'n', 'g'})
		<-c
	})
}
