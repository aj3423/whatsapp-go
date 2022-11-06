package xmpp

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Writer struct {
	Data []byte
}

func NewWriter() *Writer {
	return &Writer{Data: []byte{}}
}

func (w *Writer) push_u8(v uint8) {
	w.Data = append(w.Data, v)
}
func (w *Writer) push_u16(v uint16) {
	x := make([]byte, 2)
	binary.BigEndian.PutUint16(x, v)
	w.Data = append(w.Data, x...)
}
func (w *Writer) push_u32(v uint32) {
	x := make([]byte, 4)
	binary.BigEndian.PutUint32(x, v)
	w.Data = append(w.Data, x...)
}
func (w *Writer) push_u64(v uint64) {
	x := make([]byte, 8)
	binary.BigEndian.PutUint64(x, v)
	w.Data = append(w.Data, x...)
}
func (w *Writer) push_u20(v int) {
	x := make([]byte, 4)
	binary.BigEndian.PutUint32(x, uint32(v))
	w.Data = append(w.Data, x[1:]...)
}
func (w *Writer) write_list_start(len_ int) {
	if len_ == 0 {
		w.push_u8(LIST_EMPTY)
	} else if len_ < 256 {
		w.push_u8(LIST_8)
		w.push_u8(uint8(len_))
	} else {
		w.push_u8(LIST_16)
		w.push_u16(uint16(len_))
	}
}
func (w *Writer) write_token(token int) error {
	if token >= 0 && token <= 0xff {
		w.push_u8(uint8(token))
		return nil
	}
	return errors.New(``)
}
func (w *Writer) push_string(s string) {
	w.Data = append(w.Data, []byte(s)...)
}
func (w *Writer) write_bytes(data []byte, packed bool) {
	len_ := len(data)
	if len_ > 0x100000 {
		w.push_u8(BINARY_32)
		w.push_u32(uint32(len_))
		w.Data = append(w.Data, data...)
	} else if len_ >= 256 {
		w.push_u8(BINARY_20)
		w.push_u20(len_)
		w.Data = append(w.Data, data...)
	} else if packed {
		bs, e := pack_bytes(NIBBLE_8, data)
		if e != nil {
			bs, e = pack_bytes(HEX_8, data)
		}
		if e == nil {
			w.Data = append(w.Data, bs...)
		} else {
			w.write_raw_bytes(data)
		}
	} else {
		w.write_raw_bytes(data)
	}
}
func (w *Writer) write_raw_bytes(data []byte) {
	w.push_u8(BINARY_8)
	w.push_u8(uint8(len(data)))
	w.Data = append(w.Data, data...)
}

// user:        cc+phone
// agent, dev:  .0:1
// server:      @s.whatsapp.net
func (w *Writer) write_dev_jid_pair(
	user string, agent, device uint8, server string,
) {
	w.push_u8(DEV_JID_PAIR)
	w.push_u8(agent)
	w.push_u8(device)
	w.write_string(user, true, false)
}
func (w *Writer) write_jid(user, server string) {
	w.push_u8(JID_PAIR)
	if len(user) > 0 {
		w.write_string(user, true, false)
	} else {
		w.write_token(LIST_EMPTY)
	}
	w.write_string(server, false, false)
}

func pack_nibble(n uint8) (uint8, error) {
	if n >= '0' && n <= '9' {
		return n - '0', nil
	} else if n == '-' || n == '.' {
		return 10 + (n - 45), nil
	}
	return 0, errors.New(``)
}
func pack_hex(v uint8) (uint8, error) {
	if v >= '0' && v <= '9' {
		return v - '0', nil
	}
	if v >= 'A' && v <= 'F' {
		return v - 'A' + 0xa, nil
	}
	if v >= 'a' && v <= 'f' {
		return v - 'a' + 0xa, nil
	}
	return 0, errors.New(``)
}
func pack_byte(type_, p1, p2 uint8) (uint8, error) {
	switch type_ {
	case NIBBLE_8:
		b1, e1 := pack_nibble(p1)
		b2, e2 := pack_nibble(p2)
		if e1 != nil || e2 != nil {
			return 0, errors.New(``)
		}
		return b1<<4 | b2, nil
	case HEX_8:
		b1, e1 := pack_hex(p1)
		b2, e2 := pack_hex(p2)
		if e1 != nil || e2 != nil {
			return 0, errors.New(``)
		}
		return b1<<4 | b2, nil
	}
	return 0, errors.New(``)
}
func write_u8(ss []byte, b uint8) []byte {
	return append(ss, b)
}
func pack_bytes(type_ uint8, data []byte) ([]byte, error) {
	len_ := len(data)
	if len_ >= 128 {
		return nil, errors.New(``)
	}
	tss := []byte{}
	tss = write_u8(tss, type_)

	var b2 uint8
	if len_%2 > 0 {
		b2 = 0x80
	} else {
		b2 = 0
	}
	b2 |= uint8(math.Ceil(float64(len_) / 2))
	tss = write_u8(tss, b2)

	// [0:len-1] bytes
	for i := 0; i < len_/2; i++ {
		b, e := pack_byte(type_, data[2*i], data[2*i+1])
		if e != nil {
			return nil, e
		}
		tss = write_u8(tss, b)
	}
	// last byte
	if len_%2 != 0 {
		b, e := pack_byte(type_, data[len_-1], '0')
		if e != nil {
			return nil, e
		}
		b |= 0xF
		tss = write_u8(tss, b)
	}
	return tss, nil
}

func (w *Writer) write_string(str string, packed, try_jid bool) {
	idx, ok := MapDict_0[str]
	if ok {
		if idx.Idx_0 == 0 { // in Dict_0
			w.write_token(idx.Idx_1)
		} else { // in Dict_1
			w.write_token(idx.Idx_0)
			w.write_token(idx.Idx_1)
		}
	} else { // not found
		if try_jid {
			isCompliant := true

			user, agent, device, server, jid_type := analyze_jid(str)
			switch jid_type {
			case NormalJid: // "xxx@s.whatsapp.net"
				w.write_jid(strconv.Itoa(user), server)

			case DevJid: // "xxx.0:1@s.whatsapp.net"
				w.write_dev_jid_pair(strconv.Itoa(user), agent, device, server)

			default: // not jid
				if str == `s.whatsapp.net` {
					isCompliant = true
				} else if str == `g.us` {
					isCompliant = true
				} else {
					isCompliant = false
				}
				if isCompliant {
					w.write_jid(``, str)
				} else {
					w.write_bytes([]byte(str), packed)
				}
			}
		} else {
			w.write_bytes([]byte(str), packed)
		}
	}
}
func is_compliant(kv *KeyValue) bool {
	if kv.Type == 1 {
		return true
	}
	if strings.Contains(kv.Value, `s.whatsapp.net`) {
		return true
	}
	if strings.Contains(kv.Value, `g.us`) {
		return true
	}
	return false
}
func (w *Writer) WriteNode(n *Node) []byte {
	len_ := 1
	len_ += len(n.Attrs) * 2
	if len(n.Children) > 0 {
		len_ += 1
	}
	if len(n.Data) > 0 {
		len_ += 1
	}

	w.write_list_start(len_)

	// Tag
	w.write_string(n.Tag, false, false)

	// Attr
	for _, a := range n.Attrs {
		// write key
		w.write_string(a.Key, false, false)

		// write value
		/** apk uses
		    if(1 == attr.type && (01j.isCompliant(v_jid))) {
			    this.writeJid(ss, v_jid);
		    }
		    else {
			    this.writeString(ss, attr.value, true, true);
		    }
		*/
		// Here simply use string search.
		//if is_compliant(a) {
		//v := strings.Split(a.Value, `@`)
		//if len(v) == 1 {
		//w.write_jid(``, v[0])
		//} else {
		//w.write_jid(v[0], v[1])
		//}
		//} else {
		w.write_string(a.Value, true, true)
		//}
	}

	// Data
	if len(n.Data) > 0 {
		w.write_bytes(n.Data, false)
	}
	// Children
	if len(n.Children) > 0 {
		w.write_list_start(len(n.Children))
		for _, c := range n.Children {
			w.WriteNode(c)
		}
	}
	return w.Data
}
