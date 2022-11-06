package wam

import (
	"errors"

	"wa/crypto"
)

type ByteBuffer struct {
	Data   []byte
	offset int // offset of Data, used for parsing
}

func (bf *ByteBuffer) read_n(num int) ([]byte, error) {
	if len(bf.Data[bf.offset:]) < num {
		return nil, errors.New(`not enough bytes`)
	}
	ret := bf.Data[bf.offset : bf.offset+num]
	bf.offset += num
	return ret, nil
}
func (bf *ByteBuffer) read_u8() (uint8, error) {
	bs, e := bf.read_n(1)
	if e != nil {
		return 0, e
	}
	return crypto.LE2U8(bs), nil
}
func (bf *ByteBuffer) read_u16() (uint16, error) {
	bs, e := bf.read_n(2)
	if e != nil {
		return 0, e
	}
	return crypto.LE2U16(bs), nil
}
func (bf *ByteBuffer) read_u32() (uint32, error) {
	bs, e := bf.read_n(4)
	if e != nil {
		return 0, e
	}
	return crypto.LE2U32(bs), nil
}
func (bf *ByteBuffer) read_u64() (uint64, error) {
	bs, e := bf.read_n(8)
	if e != nil {
		return 0, e
	}
	return crypto.LE2U64(bs), nil
}
func (bf *ByteBuffer) put_byte(val uint8) {
	bf.Data = append(bf.Data, val)
}
func (bf *ByteBuffer) put_bytes(val []byte) {
	bf.Data = append(bf.Data, val...)
}

func (bf *ByteBuffer) put_number(val int) uint8 {
	if val == 0 { // bool
		return 1
	}
	if val == 1 { // bool
		return 2
	}
	if -0x80 <= val && val <= 0x7F {
		bf.Data = append(bf.Data, crypto.U82LE(uint8(val))...)
		return 3
	}
	if -0x8000 <= val && val <= 0x7FFF {
		bf.Data = append(bf.Data, crypto.U162LE(uint16(val))...)
		return 4
	}
	if -0x80000000 <= val && val <= 0x7FFFFFFF {
		bf.Data = append(bf.Data, crypto.U322LE(uint32(val))...)
		return 5
	}

	bf.Data = append(bf.Data, crypto.U642LE(uint64(val))...)
	return 6
}

func (bf *ByteBuffer) put_u32(val uint32) int {
	if val <= 0xff {
		bf.Data = append(bf.Data, crypto.U82LE(uint8(val))...)
		return 1
	}
	if val <= 0xffff {
		bf.Data = append(bf.Data, crypto.U162LE(uint16(val))...)
		return 2
	}

	bf.Data = append(bf.Data, crypto.U322LE(uint32(val))...)
	return 4
}

func (bf *ByteBuffer) read_n_bytes_as_int32(
	numBytes int,
) (int32, error) {
	var ret int32 = 0
	for i := 0; i < numBytes; i++ {
		b, e := bf.read_u8()
		if e != nil {
			return 0, e
		}
		ret |= (int32(b)) << (i << 3)
	}
	return ret, nil
}
