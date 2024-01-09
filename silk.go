package silk

import (
	"bytes"
	"encoding/binary"
	"io"
)

func Decode(reader io.Reader, writer io.Writer) (err error) {
	r := newExReader(reader)

	magic := make([]byte, len(MagicV3), len(MagicV3))
	if err = r.FullRead(magic); err != nil {
		return
	}

	if bytes.Compare(magic, MagicV3) != 0 {
		return ErrMagicNotMatch
	}

	var decoder *Decoder

	if decoder, err = NewDecoder(); err != nil {
		return err
	}

	var (
		payloadSize uint16
		out         []int16
		n           int
	)

	allocBuffer := make([]byte, MaxArithmBytes, MaxArithmBytes)
	outAllocBuffer := make([]byte, MaxApiFSKHZ*FrameLengthMS, MaxApiFSKHZ*FrameLengthMS)

	for {
		if payloadSize, err = r.ReadUInt16(); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}

		if payloadSize > MaxArithmBytes {
			return ErrRangeCodeDecodePayloadTooLarge
		}

		if err = r.FullRead(allocBuffer[:payloadSize]); err != nil {
			return
		}

		if out, err = decoder.Decode(false, allocBuffer[:payloadSize]); err != nil {
			return
		}

		for i := 0; i < len(out); i++ {
			binary.LittleEndian.PutUint16(outAllocBuffer[i*2:], uint16(out[i]))
		}

		size := len(out) * 2

		if n, err = writer.Write(outAllocBuffer[:size]); err != nil {
			return
		} else if n != size {
			return io.ErrShortWrite
		}
	}

	// TODO resample here
	return
}

func DecodeBytes(reader io.Reader) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	if err := Decode(reader, buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
