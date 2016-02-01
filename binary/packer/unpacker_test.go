package packer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPopByte(t *testing.T) {
	buf := new(bytes.Buffer)
	p := NewPacker(buf)
	u := NewUnpacker(buf)
	p.PushByte(0x01)
	b, err := u.PopByte()
	assert.Equal(t, err, nil, "Has error.")
	assert.Equal(t, b, byte(0x01), "byte error.")
}

func TestPopNByte(t *testing.T) {
	buf := new(bytes.Buffer)
	p := NewPacker(buf)
	u := NewUnpacker(buf)
	p.PushBytes([]byte{0x01, 0x02})
	bs, err := u.PopNBytes(2)
	assert.Equal(t, err, nil, "Has error.")
	assert.Equal(t, bs, []byte{0x01, 0x02}, "byte error.")
}

func TestPopUint16(t *testing.T) {
	buf := new(bytes.Buffer)
	p := NewPacker(buf)
	u := NewUnpacker(buf)
	p.PushUint16(1)
	i, err := u.PopUint16()
	assert.Equal(t, err, nil, "Has error.")
	assert.Equal(t, i, uint16(1), "uint16 error.")
}

func TestPopUint32(t *testing.T) {
	buf := new(bytes.Buffer)
	p := NewPacker(buf)
	u := NewUnpacker(buf)
	p.PushUint32(1)
	i, err := u.PopUint32()
	assert.Equal(t, err, nil, "Has error.")
	assert.Equal(t, i, uint32(1), "uint32 error.")
}

func TestPopUint64(t *testing.T) {
	buf := new(bytes.Buffer)
	p := NewPacker(buf)
	u := NewUnpacker(buf)
	p.PushUint64(1)
	i, err := u.PopUint64()
	assert.Equal(t, err, nil, "Has error.")
	assert.Equal(t, i, uint64(1), "uint64 error.")
}

func TestPopString(t *testing.T) {
	buf := new(bytes.Buffer)
	p := NewPacker(buf)
	u := NewUnpacker(buf)
	p.PushString("Hi")
	s, err := u.PopString(2)
	assert.Equal(t, err, nil, "Has error.")
	assert.Equal(t, s, "Hi", "string error.")
}
