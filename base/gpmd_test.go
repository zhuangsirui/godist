package base

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGPMD(t *testing.T) {
	convey.Convey("Init GPMD", t, func() {
		gpmd := GPMD{
			Host: "localhost",
			Port: 3312,
		}
		gpmd.Address()
	})
}
