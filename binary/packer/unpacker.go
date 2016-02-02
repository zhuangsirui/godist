package packer

import (
	"encoding/binary"
	"io"
)

type Unpacker struct {
	reader io.Reader
	endian binary.ByteOrder
	err    error
}

func NewUnpacker(reader io.Reader) *Unpacker {
	return &Unpacker{
		reader: reader,
		endian: binary.LittleEndian,
	}
}

func (u *Unpacker) Error() error {
	return u.err
}

func (u *Unpacker) PopByte() (byte, error) {
	buffer := make([]byte, 1)
	_, err := u.reader.Read(buffer)
	return buffer[0], err
}

func (u *Unpacker) ReadByte(b *byte) *Unpacker {
	return u.errFilter(func() {
		*b, u.err = u.PopByte()
	})
}

func (u *Unpacker) PopBytes(n uint64) ([]byte, error) {
	buffer := make([]byte, n)
	_, err := u.reader.Read(buffer)
	return buffer, err
}

func (u *Unpacker) ReadBytes(n uint64, bytes *[]byte) *Unpacker {
	return u.errFilter(func() {
		*bytes, u.err = u.PopBytes(n)
	})
}

func (u *Unpacker) PopUint16() (uint16, error) {
	buffer := make([]byte, 2)
	if _, err := u.reader.Read(buffer); err != nil {
		return 0, err
	} else {
		return u.endian.Uint16(buffer), nil
	}
}

func (u *Unpacker) ReadUint16(i *uint16) *Unpacker {
	return u.errFilter(func() {
		*i, u.err = u.PopUint16()
	})
}

func (u *Unpacker) PopUint32() (uint32, error) {
	buffer := make([]byte, 4)
	if _, err := u.reader.Read(buffer); err != nil {
		return 0, err
	} else {
		return u.endian.Uint32(buffer), nil
	}
}

func (u *Unpacker) ReadUint32(i *uint32) *Unpacker {
	return u.errFilter(func() {
		*i, u.err = u.PopUint32()
	})
}

func (u *Unpacker) PopUint64() (uint64, error) {
	buffer := make([]byte, 8)
	if _, err := u.reader.Read(buffer); err != nil {
		return 0, err
	} else {
		return u.endian.Uint64(buffer), nil
	}
}

func (u *Unpacker) ReadUint64(i *uint64) *Unpacker {
	return u.errFilter(func() {
		*i, u.err = u.PopUint64()
	})
}

func (u *Unpacker) PopString(n uint64) (string, error) {
	buffer := make([]byte, n)
	if _, err := u.reader.Read(buffer); err != nil {
		return "", err
	} else {
		return string(buffer), nil
	}
}

func (u *Unpacker) ReadString(n uint64, s *string) *Unpacker {
	return u.errFilter(func() {
		*s, u.err = u.PopString(n)
	})
}

func (u *Unpacker) errFilter(f func()) *Unpacker {
	if u.err == nil {
		f()
	}
	return u
}

func (u *Unpacker) StringWithUint16Perfix(s *string) *Unpacker {
	return u.errFilter(func() {
		var n uint16
		n, u.err = u.PopUint16()
		u.ReadString(uint64(n), s)
	})
}

func (u *Unpacker) StringWithUint32Perfix(s *string) *Unpacker {
	return u.errFilter(func() {
		var n uint32
		n, u.err = u.PopUint32()
		u.ReadString(uint64(n), s)
	})
}

func (u *Unpacker) StringWithUint64Perfix(s *string) *Unpacker {
	return u.errFilter(func() {
		var n uint64
		n, u.err = u.PopUint64()
		u.ReadString(n, s)
	})
}

func (u *Unpacker) BytesWithUint16Perfix(bytes *[]byte) *Unpacker {
	return u.errFilter(func() {
		var n uint16
		n, u.err = u.PopUint16()
		u.ReadBytes(uint64(n), bytes)
	})
}

func (u *Unpacker) BytesWithUint32Perfix(bytes *[]byte) *Unpacker {
	return u.errFilter(func() {
		var n uint32
		n, u.err = u.PopUint32()
		u.ReadBytes(uint64(n), bytes)
	})
}

func (u *Unpacker) BytesWithUint64Perfix(bytes *[]byte) *Unpacker {
	return u.errFilter(func() {
		var n uint64
		n, u.err = u.PopUint64()
		u.ReadBytes(n, bytes)
	})
}
