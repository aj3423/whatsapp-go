package xmpp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

type Reader struct {
	ss *bytes.Reader
}

func NewReader(data []byte) *Reader {
	return &Reader{
		ss: bytes.NewReader(data),
	}
}
func (r *Reader) read_list_size(tag uint8) (int, error) {
	switch tag {
	case LIST_EMPTY: // 0
		return 0, nil
	case LIST_8: // 248 (0xF8)
		r, e := r.read_u8()
		return int(r), e
	case LIST_16: // 249 (0xF9)
		r, e := r.read_u16()
		return int(r), e
	}
	return 0, errors.New("read_list_size")
}

func (r *Reader) read_u8() (uint8, error) {
	b := make([]byte, 1)
	i, e := r.ss.Read(b)
	if i != 1 || e != nil {
		return 0, errors.Wrap(e, "read_u8")
	}
	return b[0], nil
}
func (r *Reader) read_u16() (uint16, error) {
	b := make([]byte, 2)
	i, e := r.ss.Read(b)
	if i != 2 || e != nil {
		return 0, errors.Wrap(e, "read_u16")
	}
	return binary.BigEndian.Uint16(b), nil
}
func (r *Reader) read_u32() (uint32, error) {
	b := make([]byte, 4)
	i, e := r.ss.Read(b)
	if i != 4 || e != nil {
		return 0, errors.Wrap(e, "read_u32")
	}
	return binary.BigEndian.Uint32(b), nil
}
func (r *Reader) read_u64() (uint64, error) {
	b := make([]byte, 8)
	i, e := r.ss.Read(b)
	if i != 8 || e != nil {
		return 0, errors.Wrap(e, "read_u64")
	}
	return binary.BigEndian.Uint64(b), nil
}
func (r *Reader) read_u20() (int, error) {
	b := make([]byte, 3)
	i, e := r.ss.Read(b)
	if i != 3 || e != nil {
		return 0, errors.Wrap(e, "read_u20")
	}

	return (int(b[0]) << 16) + (int(b[1]) << 8) + int(b[2]), nil
}
func (r *Reader) read_packed8(tag uint8) ([]byte, error) {
	limit, e := r.read_u8()
	if e != nil {
		return nil, e
	}
	ss := []byte{}
	for i := 0; i < (int(limit) & 127); i++ {
		b, e := r.read_u8()
		if e != nil {
			return nil, e
		}
		b1, e := r.unpack_byte(tag, (b&0xf0)>>4)
		if e != nil {
			return nil, e
		}
		b2, e := r.unpack_byte(tag, b&0x0f)
		if e != nil {
			return nil, e
		}
		ss = append(ss, b1)
		ss = append(ss, b2)
	}
	if limit>>7 != 0 {
		if len(ss) > 0 {
			ss = ss[:len(ss)-1] // remove last
		}
	}
	return ss, nil
}
func (r *Reader) unpack_byte(tag, value uint8) (byte, error) {
	switch tag {
	case NIBBLE_8:
		return r.unpack_nibble(value)
	case HEX_8:
		return r.unpack_hex(value)
	}
	return 0, errors.New(`unsupported read_byte`)
}
func (r *Reader) unpack_hex(value uint8) (byte, error) {
	if value > 15 {
		return 0, errors.New(`unpack_hex`)
	}
	if value < 10 {
		return value + '0', nil
	} else {
		return value - 10 + 'A', nil
	}
}
func (r *Reader) unpack_nibble(value uint8) (byte, error) {
	if value >= 0 && value <= 9 {
		return value + '0', nil
	} else {
		switch value {
		case 10:
			return '-', nil
		case 11:
			return '.', nil
		case 15:
			return 0, nil
		}
	}
	return 0, errors.New(`upack_nibble`)
}
func (r *Reader) is_list_tag(tag uint8) bool {
	return tag == LIST_EMPTY || tag == LIST_8 || tag == LIST_16
}

func (r *Reader) read_string(tag uint8) (string, error) {
	if tag > 2 && int(tag) < len(Dict_0) {
		token, e := r.get_token(int(tag))
		if e != nil {
			return ``, e
		}
		if token == `s.whatsapp.net` {
			//token = `c.us`
		}
		return token, nil
	}
	switch tag {
	case DICTIONARY_0, DICTIONARY_1, DICTIONARY_2, DICTIONARY_3:
		// 236 ~ 239 (0xEC ~ 0xEF)
		idx_1 := tag - uint8(len(Dict_0))
		idx_2, e := r.read_u8()
		if e != nil {
			return ``, e
		}
		if int(idx_1) >= len(Dict_1) || int(idx_2) >= len(Dict_1[idx_1]) {
			return ``, errors.New(``)
		}
		return Dict_1[idx_1][idx_2], nil
	case LIST_EMPTY: // 0
		return ``, nil
	case BINARY_8: // 252 (0xFC)
		l, e := r.read_u8()
		if e != nil {
			return ``, e
		}
		bs, e := r.read_bytes(int(l))
		if e != nil {
			return ``, e
		}
		return string(bs), nil
	case BINARY_20: // 253 (0xFD)
		l, e := r.read_u20()
		if e != nil {
			return ``, e
		}
		bs, e := r.read_bytes(l)
		if e != nil {
			return ``, e
		}
		return string(bs), nil
	case BINARY_32: // 254 (0xFE)
		l, e := r.read_u32()
		if e != nil {
			return ``, e
		}
		bs, e := r.read_bytes(int(l))
		if e != nil {
			return ``, e
		}
		return string(bs), nil
	case DEV_JID_PAIR: // 247 (0xF7)
		// 918955576704.0:1@s.whatsapp.net
		agent, e := r.read_u8()
		if e != nil {
			return ``, e
		}
		device, e := r.read_u8()
		if e != nil {
			return ``, e
		}
		l, e := r.read_u8()
		if e != nil {
			return ``, e
		}
		user, e := r.read_string(l)
		if e != nil {
			return ``, e
		}
		return fmt.Sprintf("%s.%d:%d@s.whatsapp.net", user, agent, device), nil

	case JID_PAIR: // 250 (0xFA)
		l, e := r.read_u8()
		if e != nil {
			return ``, e
		}
		user, e := r.read_string(l)
		if e != nil {
			return ``, e
		}
		l, e = r.read_u8()
		if e != nil {
			return ``, e
		}
		server, e := r.read_string(l)
		if e != nil {
			return ``, e
		}
		if len(user) == 0 {
			return server, nil
		} else {
			return user + `@` + server, nil
		}
	case NIBBLE_8, HEX_8: // 255(0xFF), 251 (0xFB)
		p, e := r.read_packed8(tag)
		if e != nil {
			return ``, e
		}
		return string(p), nil
	}
	return ``, errors.New(``)
}

func (r *Reader) get_token(index int) (string, error) {
	if index < 3 || index > len(Dict_0) {
		return ``, errors.New(``)
	}
	return Dict_0[index], nil
}
func (r *Reader) read_bytes(size int) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	i, e := r.ss.Read(buf)
	if i != size || e != nil {
		return nil, errors.New(``)
	}
	return buf, nil
}
func (r *Reader) read_attributes(size int) ([]*KeyValue, error) {
	ret := []*KeyValue{}
	for i := 0; i < size; i++ {

		key_idx, e := r.read_u8()
		if e != nil {
			return nil, e
		}
		key, e := r.read_string(key_idx)
		if e != nil {
			return nil, e
		}
		val_idx, e := r.read_u8()
		if e != nil {
			return nil, e
		}
		val, e := r.read_string(val_idx)
		if e != nil {
			return nil, e
		}
		ret = append(ret, &KeyValue{Key: key, Value: val})
	}

	return ret, nil
}
func (r *Reader) read_node_list(token uint8) ([]*Node, error) {
	size, e := r.read_list_size(token)
	if e != nil {
		return nil, e
	}
	ret := []*Node{}
	for i := 0; i < size; i++ {
		n, e := r.ReadNode()
		if e != nil {
			return nil, e
		}
		ret = append(ret, n)
	}

	return ret, nil
}
func (r *Reader) ReadNode() (*Node, error) {
	b, e := r.read_u8()
	if e != nil {
		return nil, e
	}
	list_size, e := r.read_list_size(b)
	if e != nil {
		return nil, e
	}
	token, e := r.read_u8()
	if e != nil {
		return nil, e
	}
	if token == 1 {
		token, e = r.read_u8()
		if e != nil {
			return nil, e
		}
	}
	if token == STREAM_END {
		return nil, errors.New(`stream end`)
	}
	tag, e := r.read_string(token)
	if e != nil {
		return nil, e
	}
	if list_size == 0 || len(tag) == 0 {
		return nil, errors.New(``)
	}
	attrs, e := r.read_attributes((list_size - 1) / 2)
	if list_size%2 == 1 {
		return &Node{Tag: tag, Attrs: attrs}, nil
	}
	read2, e := r.read_u8()
	if e != nil {
		return nil, e
	}
	if r.is_list_tag(read2) {
		ch, e := r.read_node_list(read2)
		if e != nil {
			return nil, e
		}
		return &Node{Tag: tag, Attrs: attrs, Children: ch}, nil
	}
	switch read2 {
	case BINARY_8:
		sz, e := r.read_u8()
		if e != nil {
			return nil, e
		}
		data, e := r.read_bytes(int(sz))
		if e != nil {
			return nil, e
		}
		return &Node{Tag: tag, Attrs: attrs, Data: data}, nil
	case BINARY_20:
		sz, e := r.read_u20()
		if e != nil {
			return nil, e
		}
		data, e := r.read_bytes(int(sz))
		if e != nil {
			return nil, e
		}
		return &Node{Tag: tag, Attrs: attrs, Data: data}, nil
	case BINARY_32:
		sz, e := r.read_u32()
		if e != nil {
			return nil, e
		}
		data, e := r.read_bytes(int(sz))
		if e != nil {
			return nil, e
		}
		return &Node{Tag: tag, Attrs: attrs, Data: data}, nil
	case HEX_8, NIBBLE_8:
		data, e := r.read_packed8(read2)
		if e != nil {
			return nil, e
		}
		return &Node{Tag: tag, Attrs: attrs, Data: data}, nil
	default:
		data, e := r.read_string(read2)
		if e != nil {
			return nil, e
		}
		return &Node{Tag: tag, Attrs: attrs, Data: []byte(data)}, nil
	}
}
