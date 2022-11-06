package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"afs"
	"ahex"
	"algo"
	"wa/xmpp"
)

func bytes_2_node(b []byte) (*xmpp.Node, error) {
	if len(b) < 2 {
		return nil, errors.New(`data too short: ` + ahex.Enc(b))
	}
	if b[0] == 0 && b[1] == 0xf8 {
		b = b[1:]
	} else if b[0] == 2 && b[1] == 0x78 && b[2] == 0x9c {
		dec, e := algo.UnZlib(b[1:])
		if e != nil {
			return nil, e
		}
		return bytes_2_node(dec)
	}
	r := xmpp.NewReader(b)

	return r.ReadNode()
}

func init() {
	flag.Parse()
}

func cmd_line_mode() {
	strs := []string{}
	for i := 0; i < flag.NArg(); i++ {
		strs = append(strs, flag.Arg(i))
	}

	arg := strings.Join(strs, "")
	arg = strings.ReplaceAll(arg, "\n", "")
	arg = strings.ReplaceAll(arg, "\r", "")
	arg = strings.ReplaceAll(arg, "\t", "")
	arg = strings.ReplaceAll(arg, " ", "")

	if strings.HasPrefix(arg, "data:") {
		arg = arg[5:]
	}
	if strings.HasPrefix(arg, "result:") {
		arg = arg[7:]
	}

	n, e := bytes_2_node(ahex.Dec(arg))
	if e != nil {
		panic(e)
	}
	fmt.Println(n.ToString())
}
func file_mode() {
	lines := afs.ReadLines(flag.Arg(0))
	new_lines := []string{}
	var prev_line string
	for _, line := range lines {
		new_lines = append(new_lines, line)
		if prev_line == "---- AesGcmWrap.Decrypt: ----" && strings.HasPrefix(line, "result:") {
			if n, e := bytes_2_node(ahex.Dec(line[7:])); e == nil {
				new_lines = append(new_lines, n.ToString())
			}
		} else if prev_line == "---- AesGcmWrap.Encrypt: ----" && strings.HasPrefix(line, "data:") {
			if n, e := bytes_2_node(ahex.Dec(line[5:])); e == nil {
				new_lines = append(new_lines, n.ToString())
			}
		}

		prev_line = line
	}

	f := strings.Join(new_lines, "\n")
	afs.Write(flag.Arg(0), []byte(f))
}

func main() {
	if afs.Exist(flag.Arg(0)) {
		file_mode()
	} else {
		cmd_line_mode()
	}
}
