package gpmd

import (
	"bytes"
	"fmt"
	"godist/base"
	"net"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/zhuangsirui/binpacker"
)

const (
	testPort = 1989
	testHost = "localhost"
)

func TestInit(t *testing.T) {
	convey.Convey("Init", t, func() {
		convey.Convey("Init should panic resolve addr", func() {
			SetPort(22)
			SetHost("xx8*10123-=")
			convey.So(func() {
				Init()
			}, convey.ShouldPanic)
		})
		convey.Convey("Init should panic for listen", func() {
			SetPort(22)
			SetHost("google.com")
			convey.So(func() {
				Init()
			}, convey.ShouldPanic)
		})
		convey.Convey("Init success", func() {
			SetPort(testPort)
			SetHost(testHost)
			convey.So(func() {
				Init()
			}, convey.ShouldNotPanic)
		})
	})
}

func TestRegistion(t *testing.T) {
	convey.Convey("Registion", t, func() {
		requestBuf := new(bytes.Buffer)
		binpacker.NewPacker(requestBuf).
			PushByte(REQ_REGISTER).
			PushUint16(20130).
			PushUint16(uint16(len("agent_name"))).
			PushString("agent_name").
			PushUint16(uint16(len(testHost))).
			PushString(testHost)
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
		convey.So(err, convey.ShouldBeNil)
		conn, err := net.DialTCP("tcp", nil, address)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("Register new node", func() {
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_REGISTER, ACK_RES_OK})
		})

		convey.Convey("Register an exist node", func() {
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_REGISTER, ACK_RES_NODE_EXIST})
		})

		requestBuf.Reset()
		binpacker.NewPacker(requestBuf).
			PushByte(REQ_UNREGISTER).
			PushUint16(uint16(len("agent_name"))).
			PushString("agent_name")

		convey.Convey("Unregister an exist node", func() {
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			t.Log(ack)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_UNREGISTER, ACK_RES_OK})
		})

		convey.Convey("Unregister an not exist node", func() {
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			t.Log(ack)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_UNREGISTER, ACK_RES_NODE_NOT_EXIST})
		})
	})
}

func TestQuery(t *testing.T) {
	convey.Convey("Query", t, func() {
		node := &base.Node{
			Port: 8899,
			Host: "localhost",
			Name: "test_query_node",
		}
		requestBuf := new(bytes.Buffer)
		binpacker.NewPacker(requestBuf).
			PushByte(REQ_QUERY).
			PushUint16(uint16(len(node.Name))).
			PushString(node.Name)
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", testHost, testPort))
		convey.So(err, convey.ShouldBeNil)
		conn, err := net.DialTCP("tcp", nil, address)
		convey.So(err, convey.ShouldBeNil)

		_manager.register(node)
		convey.Convey("Query an exist node", func() {
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_QUERY, ACK_RES_OK})
		})

		_manager.unregister(node.Name)
		convey.Convey("Query a not exist node", func() {
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_QUERY, ACK_RES_NODE_NOT_EXIST})
		})
	})
}

func TestRestart(t *testing.T) {
	convey.Convey("Restart", t, func() {
		_manager.listener.Close()
		time.Sleep(50 * time.Millisecond)
	})
}

func TestStop(t *testing.T) {
	convey.Convey("Restart", t, func() {
		convey.So(func() {
			Stop()
		}, convey.ShouldNotPanic)
	})
}
