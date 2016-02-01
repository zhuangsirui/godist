package packer

import (
	"encoding/binary"
	"io"
)

type Unpacker struct {
	reader io.Reader
	endian binary.ByteOrder
}

func NewUnpacker(reader io.Reader) *Unpacker {
	return &Unpacker{
		reader: reader,
		endian: binary.LittleEndian,
	}
}

func (u *Unpacker) PopByte() (byte, error) {
	buffer := make([]byte, 1)
	_, err := u.reader.Read(buffer)
	return buffer[0], err
}

func (u *Unpacker) PopNBytes(n uint64) ([]byte, error) {
	buffer := make([]byte, n)
	_, err := u.reader.Read(buffer)
	return buffer, err
}

func (u *Unpacker) PopUint16() (uint16, error) {
	buffer := make([]byte, 2)
	if _, err := u.reader.Read(buffer); err != nil {
		return 0, err
	} else {
		return u.endian.Uint16(buffer), nil
	}
}

func (u *Unpacker) PopUint32() (uint32, error) {
	buffer := make([]byte, 4)
	if _, err := u.reader.Read(buffer); err != nil {
		return 0, err
	} else {
		return u.endian.Uint32(buffer), nil
	}
}

func (u *Unpacker) PopUint64() (uint64, error) {
	buffer := make([]byte, 8)
	if _, err := u.reader.Read(buffer); err != nil {
		return 0, err
	} else {
		return u.endian.Uint64(buffer), nil
	}
}

func (u *Unpacker) PopString(n uint64) (string, error) {
	buffer := make([]byte, n)
	if _, err := u.reader.Read(buffer); err != nil {
		return "", err
	} else {
		return string(buffer), nil
	}
}
