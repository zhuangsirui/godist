package godist

import (
	"bytes"
	"testing"
)

var tAgent6 *Agent

var nodeName6, nodeName7 = "testnode6@localhost", "testnode7@localhost"

func TestNewProcess(t *testing.T) {
	tAgent6 = New(nodeName6)
	process := tAgent6.NewProcess()
	process.GetId()
	replyChann := make(chan []byte)
	process.Run(func(message []byte) {
		if bytes.Compare(message, []byte{'p', 'i', 'n', 'g'}) != 0 {
			t.Error("message error")
		}
		replyChann <- []byte{'p', 'o', 'n', 'g'}
	})
	process.Channel <- []byte{'p', 'i', 'n', 'g'}
	reply := <-replyChann
	if bytes.Compare(reply, []byte{'p', 'o', 'n', 'g'}) != 0 {
		t.Error("message error")
	}
}

func TestStaticNewProcess(t *testing.T) {
	Init(nodeName7)
	process := NewProcess()
	replyChann := make(chan []byte)
	process.Run(func(message []byte) {
		if bytes.Compare(message, []byte{'p', 'i', 'n', 'g'}) != 0 {
			t.Error("message error")
		}
		replyChann <- []byte{'p', 'o', 'n', 'g'}
	})
	process.Channel <- []byte{'p', 'i', 'n', 'g'}
	reply := <-replyChann
	if bytes.Compare(reply, []byte{'p', 'o', 'n', 'g'}) != 0 {
		t.Error("message error")
	}
}
