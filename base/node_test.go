package base

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNode(t *testing.T) {
	convey.Convey("Init Node", t, func() {
		n := Node{
			Port: 3312,
			Host: "localhost",
			Name: "master_01",
		}
		n.FullName()
		n.String()
	})
}
