package wam

import (
	"bytes"
	"errors"
)

var Head = []byte{
	0x57, 0x41, 0x4d, 0x05,
}

func Parse(
	bs []byte,
) ([]*Record, error) {
	if len(bs) < len(Head) || !bytes.Equal(bs[0:len(Head)], Head) {
		return nil, errors.New(`less8`)
	}
	offset := 8 // skip header

	ret := []*Record{}
	rest := len(bs) - offset

	for rest > 0 {
		rec := &Record{}
		used_len, e := rec.Parse(bs[offset:])
		if e != nil {
			return nil, e
		}
		ret = append(ret, rec)

		rest -= used_len
		offset += used_len
	}

	return ret, nil
}
