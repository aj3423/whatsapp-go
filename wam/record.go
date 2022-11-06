package wam

import (
	"errors"
	"fmt"
)

// Record.ClassType
/*
const (
	ClassType_Wild  = 0
	ClassType_Begin = 1
	ClassType_Data  = 2
)
*/

type Record struct {
	ClassType uint8
	Id        int32
	Value     interface{}

	IsClassEnd bool // id | 4
}

func (rec *Record) ToBytes() ([]byte, error) {
	w := ByteBuffer{}
	first := 0
	w.put_byte(0) // placeholder for id-type

	id_len := w.put_u32(uint32(rec.Id))
	if id_len > 2 { // more than 2 bytes
		return nil, errors.New(`id too big to fit in 2 bytes`)
	}
	id_is_two_bytes := uint8(0)
	if id_len == 2 {
		id_is_two_bytes = 1
	}

	var val_type uint8
	if rec.Value == nil {
		val_type = 0
	} else if v, ok := rec.Value.(bool); ok {
		if v {
			val_type = w.put_number(1)
		} else {
			val_type = w.put_number(0)
		}

	} else if v, ok := rec.Value.(int); ok {
		val_type = w.put_number(v)
	} else if v, ok := rec.Value.(int64); ok {
		val_type = w.put_number(int(v))
	} else if v, ok := rec.Value.(int32); ok {
		val_type = w.put_number(int(v))
	} else if v, ok := rec.Value.(float64); ok {
		v2 := int(v)
		if float64(v2) == v {
			val_type = w.put_number(int(v))
		} else {
			return nil, errors.New(`doubleToRawLongBits not implemented`)
		}
	} else if v, ok := rec.Value.(string); ok {
		len_ := len(v)
		if len_ > 0x400 {
			return nil, errors.New(`limited to 0x400 bytes`)
		}
		len_len := w.put_u32(uint32(len_))
		w.put_bytes([]byte(v))

		switch len_len {
		case 1:
			val_type = 8
		case 2:
			val_type = 9
		case 4:
			val_type = 10
		default:
			return nil, errors.New(`impossible`)
		}
	} else {
		return nil, errors.New(`wtf val_type: ` + fmt.Sprintf("%v", rec.Value))
	}

	w.Data[first] = rec.ClassType | (val_type<<4 | (id_is_two_bytes)<<3)
	return w.Data, nil
}

func (rec *Record) Parse(
	data []byte,
) (int, error) {

	bf := ByteBuffer{Data: data}

	b1, e := bf.read_u8()
	if e != nil {
		return 0, e
	}
	class_type := b1 & 3
	if class_type > 2 {
		return 0, errors.New(`Invalid type`)
	}
	id_is_two_bytes := 1
	if (b1 & 8) == 0 {
		id_is_two_bytes = 0
	}
	var id int32
	if id_is_two_bytes == 0 {
		id, e = bf.read_n_bytes_as_int32(1)
	} else {
		id, e = bf.read_n_bytes_as_int32(2)
	}
	if e != nil {
		return 0, e
	}
	val_type := b1 >> 4 & 15
	if val_type > 10 {
		return 0, errors.New(`valt > 10`)
	}

	rec.ClassType = class_type
	rec.Id = id
	if b1&4 == 4 {
		rec.IsClassEnd = true
	} else {
		rec.IsClassEnd = false
	}

	switch val_type {
	case 0:
		rec.Value = nil
		return bf.offset, nil
	case 1:
		rec.Value = int32(0)
		return bf.offset, nil
	case 2:
		rec.Value = int32(1)
		return bf.offset, nil
	case 3:
		if b, e := bf.read_u8(); e == nil {
			rec.Value = int8(b)
			return bf.offset, nil
		}
	case 4:
		if b, e := bf.read_u16(); e == nil {
			rec.Value = int16(b)
			return bf.offset, nil
		}
	case 5:
		if b, e := bf.read_u32(); e == nil {
			rec.Value = int32(b)
			return bf.offset, nil
		}
	case 6:
		if b, e := bf.read_u64(); e == nil {
			rec.Value = int64(b)
			return bf.offset, nil
		}
	case 7:
		if b, e := bf.read_u64(); e == nil {
			rec.Value = float64(b)
			return bf.offset, nil
		}
	case 8:
		if len_, e := bf.read_n_bytes_as_int32(1); e == nil {
			if bs, e := bf.read_n(int(len_)); e == nil {
				rec.Value = string(bs)
				return bf.offset, nil
			}
		}
	case 9:
		if len_, e := bf.read_n_bytes_as_int32(2); e == nil {
			if bs, e := bf.read_n(int(len_)); e == nil {
				rec.Value = string(bs)
				return bf.offset, nil
			}
		}
	case 10:
		if len_, e := bf.read_n_bytes_as_int32(4); e == nil {
			if bs, e := bf.read_n(int(len_)); e == nil {
				rec.Value = string(bs)
				return bf.offset, nil
			}
		}
	}

	return 0, errors.New(`wtf err`)
}
