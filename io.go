package silk

import (
	"encoding/binary"
	"io"
	"unsafe"
)

const (
	UInt8Size  = int(unsafe.Sizeof(uint8(0)))
	UInt16Size = int(unsafe.Sizeof(uint16(0)))
	UInt32Size = int(unsafe.Sizeof(uint32(0)))
	UInt64Size = int(unsafe.Sizeof(uint64(0)))
)

// exReader not thread safe!!
type exReader struct {
	io.Reader
	buf   []byte
	order binary.ByteOrder
}

func newExReader(r io.Reader) *exReader {
	return &exReader{
		Reader: r,
		order:  binary.LittleEndian,
		buf:    make([]byte, UInt64Size, UInt64Size),
	}
}

func (er *exReader) readBuf(size int) ([]byte, error) {
	if err := er.FullRead(er.buf[:size]); err != nil {
		return nil, err
	}
	return er.buf[:size], nil
}

func (er *exReader) FullRead(buf []byte) error {
	n, err := er.Read(buf)
	if err != nil {
		return err
	}
	if n < len(buf) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (er *exReader) ReadByte() (byte, error) {
	if buf, err := er.readBuf(UInt8Size); err != nil {
		return 0, err
	} else {
		return buf[0], nil
	}
}

func (er *exReader) ReadUInt16() (uint16, error) {
	if buf, err := er.readBuf(UInt16Size); err != nil {
		return 0, err
	} else {
		return er.order.Uint16(buf), nil
	}
}

func (er *exReader) ReadUInt32() (uint32, error) {
	if buf, err := er.readBuf(UInt32Size); err != nil {
		return 0, err
	} else {
		return er.order.Uint32(buf), nil
	}
}

func (er *exReader) ReadUInt64() (uint64, error) {
	if buf, err := er.readBuf(UInt64Size); err != nil {
		return 0, err
	} else {
		return er.order.Uint64(buf), nil
	}
}

func memset[T any](a []T, v T, s int) {
	for i := 0; i < s; i++ {
		a[i] = v
	}
}

func memcpy[T any](a, b []T, s int) {
	for i := 0; i < s; i++ {
		a[i] = b[i]
	}
}
