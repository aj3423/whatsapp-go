package wam

import "bytes"

type Chunk interface {
	Append(*Record)
}

// Wild
type WildChunk struct {
	Records []*Record
}

func (c *WildChunk) Append(id int32, val interface{}) {
	c.Records = append(c.Records, &Record{
		ClassType: 0,
		Id:        id,
		Value:     val,
	})
}

func (c *WildChunk) ToBytes() ([]byte, error) {
	bs := []byte{}
	for _, rec := range c.Records {
		b, e := rec.ToBytes()
		if e != nil {
			return nil, e
		}
		bs = append(bs, b...)
	}
	return bs, nil
}

// Class
type ClassChunk struct {
	Id      int32
	Value   interface{}
	Records []*Record
}

func (c *ClassChunk) Append(id int32, val interface{}) {
	c.Records = append(c.Records, &Record{
		ClassType: 2,
		Id:        id,
		Value:     val,
	})
}

func (c *ClassChunk) ToBytes() ([]byte, error) {
	bs := [][]byte{}
	// append first class id rec
	frec := Record{
		ClassType: 1,
		Id:        c.Id,
		Value:     c.Value,
	}
	b, e := frec.ToBytes()
	if e != nil {
		return nil, e
	}
	bs = append(bs, b)
	for _, rec := range c.Records {
		b, e := rec.ToBytes()
		if e != nil {
			return nil, e
		}
		bs = append(bs, b)
	}
	last := bs[len(bs)-1]
	last[0] |= 4
	return bytes.Join(bs, []byte{}), nil
}
