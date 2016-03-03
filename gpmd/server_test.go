package gpmd

import (
	"bytes"
	"fmt"
	"godist/base"
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
			SetPort(1989)
			SetHost("localhost")
			convey.So(func() {
				Init()
			}, convey.ShouldNotPanic)
			Started()
			Stop()
			Stopped()
		})
	})
}

func TestServer(t *testing.T) {
	convey.Convey("Server", t, func() {
		SetPort(1989)
		Init()
		Started()
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", Host(), Port()))
		convey.So(err, convey.ShouldBeNil)
		requestBuf := new(bytes.Buffer)

		convey.Convey("Register", func() {
			binpacker.NewPacker(requestBuf).
				PushByte(REQ_REGISTER).
				PushUint16(20130).
				PushUint16(uint16(len("agent_name"))).
				PushString("agent_name").
				PushUint16(uint16(len(Host()))).
				PushString(Host())

			convey.Convey("Register new", func() {
				_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
				convey.So(err, convey.ShouldBeNil)
				unpacker := binpacker.NewUnpacker(conn)
				var ackCode, ackState byte
				unpacker.FetchByte(&ackCode).FetchByte(&ackState)
				convey.So(ackCode, convey.ShouldEqual, REQ_REGISTER)
				convey.So(ackState, convey.ShouldEqual, ACK_RES_OK)
			})

			convey.Convey("Register old", func() {
				_manager.register(&base.Node{
					Name: "agent_name",
					Port: 1999,
				})
				_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
				convey.So(err, convey.ShouldBeNil)
				var ackCode, ackState byte
				binpacker.NewUnpacker(conn).FetchByte(&ackCode).FetchByte(&ackState)
				convey.So(ackCode, convey.ShouldEqual, REQ_REGISTER)
				convey.So(ackState, convey.ShouldEqual, ACK_RES_NODE_EXIST)
			})
		})

		convey.Convey("Unregister", func() {
			binpacker.NewPacker(requestBuf).
				PushByte(REQ_UNREGISTER).
				PushUint16(uint16(len("agent_name"))).
				PushString("agent_name")

			convey.Convey("Unregister exist", func() {
				_manager.register(&base.Node{
					Name: "agent_name",
					Port: 1999,
				})
				_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
				convey.So(err, convey.ShouldBeNil)
				var ackCode, ackState byte
				binpacker.NewUnpacker(conn).FetchByte(&ackCode).FetchByte(&ackState)
				convey.So(ackCode, convey.ShouldEqual, REQ_UNREGISTER)
				convey.So(ackState, convey.ShouldEqual, ACK_RES_OK)
			})

			convey.Convey("Unregister not exist", func() {
				_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
				convey.So(err, convey.ShouldBeNil)
				var ackCode, ackState byte
				binpacker.NewUnpacker(conn).FetchByte(&ackCode).FetchByte(&ackState)
				convey.So(ackCode, convey.ShouldEqual, REQ_UNREGISTER)
				convey.So(ackState, convey.ShouldEqual, ACK_RES_NODE_NOT_EXIST)
			})
		})

		convey.Convey("Send unknown code", func() {
			binpacker.NewPacker(requestBuf).PushByte(0xff)
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
		})

		Stop()
		Stopped()
	})
}

func TestQuery(t *testing.T) {
	convey.Convey("Query", t, func() {
		SetPort(1989)
		SetHost("")
		Init()
		Started()

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
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", Host(), Port()))
		convey.So(err, convey.ShouldBeNil)
		conn, err := net.DialTCP("tcp", nil, address)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("Query an exist node", func() {
			_manager.register(node)
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_QUERY, ACK_RES_OK})
			_manager.unregister(node.Name)
		})

		convey.Convey("Query a not exist node", func() {
			_manager.unregister(node.Name)
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
			ack := make([]byte, 2)
			_, err = conn.Read(ack)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ack, convey.ShouldResemble, []byte{REQ_QUERY, ACK_RES_NODE_NOT_EXIST})
		})

		Stop()
		Stopped()
	})
}

func TestRestart(t *testing.T) {
	convey.Convey("Restart", t, func() {
		SetPort(1989)
		SetHost("")
		Init()
		Started()
		_manager.listener.Close()
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", Host(), Port()))
		convey.So(err, convey.ShouldBeNil)
		var restarted bool
		for !restarted {
			if _, err := net.DialTCP("tcp", nil, address); err == nil {
				restarted = true
			}
		}
		Stop()
		Stopped()
	})
}

func TestStop(t *testing.T) {
	convey.Convey("Stop", t, func() {
		SetPort(1989)
		Init()
		Started()
		convey.So(func() {
			Stop()
		}, convey.ShouldNotPanic)
		Stopped()
	})
}

/*
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
			Stop()
			Stopped()
		})
	})
}

func TestRegistion(t *testing.T) {
	convey.Convey("Registion", t, func() {
		SetPort(testPort)
		SetHost(testHost)
		Init()
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

			convey.Convey("Register an exist node", func() {
				_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
				convey.So(err, convey.ShouldBeNil)
				ack := make([]byte, 2)
				_, err = conn.Read(ack)
				convey.So(err, convey.ShouldBeNil)
				convey.So(ack, convey.ShouldResemble, []byte{REQ_REGISTER, ACK_RES_NODE_EXIST})

				convey.Convey("Unregister an exist node", func() {
					requestBuf.Reset()
					binpacker.NewPacker(requestBuf).
						PushByte(REQ_UNREGISTER).
						PushUint16(uint16(len("agent_name"))).
						PushString("agent_name")
					_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
					convey.So(err, convey.ShouldBeNil)
					ack := make([]byte, 2)
					_, err = conn.Read(ack)
					convey.So(err, convey.ShouldBeNil)
					t.Log(ack)
					convey.So(ack, convey.ShouldResemble, []byte{REQ_UNREGISTER, ACK_RES_OK})

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
			})
		})

		convey.Convey("Send unknown code", func() {
			requestBuf := new(bytes.Buffer)
			binpacker.NewPacker(requestBuf).PushByte(0xff)
			_, err = conn.Write(binpacker.AddUint16Perfix(requestBuf.Bytes()))
			convey.So(err, convey.ShouldBeNil)
		})

		Stop()
		Stopped()
	})
}

/*

func TestRestart(t *testing.T) {
	convey.Convey("Restart", t, func() {
		SetPort(testPort)
		SetHost(testHost)
		Init()
		_manager.listener.Close()
		_manager.Restarted()
		Stop()
		Stopped()
	})
}

func TestStop(t *testing.T) {
	convey.Convey("Stop", t, func() {
		SetPort(testPort)
		SetHost(testHost)
		Init()
		convey.So(func() {
			Stop()
		}, convey.ShouldNotPanic)
		Stopped()
	})
}
*/
