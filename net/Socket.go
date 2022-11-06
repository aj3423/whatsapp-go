package net

import (
	"encoding/binary"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
	"wa/def"
)

//var Sock_Port = "443"

var Sock_Port = "5222"

type Socket struct {
	net.Conn
	Proxy string
	Dns   map[string]string

	wg    sync.WaitGroup
	rLock sync.Mutex //deadlock.Mutex
	wLock sync.Mutex //deadlock.Mutex
}

func (this *Socket) Connect() error {
	var e error

	var wa_host string = "g.whatsapp.net"

	if this.Dns != nil {
		ip, ok := this.Dns[wa_host]
		if ok { // find Ip in dns mapping
			wa_host = ip
		}
	}

	if len(this.Proxy) > 0 {
		u, e := url.Parse(this.Proxy)
		if e != nil {
			return errors.Wrap(e, `create socks5 fail`)
		}

		dialer, e := proxy.FromURL(u, &net.Dialer{
			Timeout: time.Duration(def.NET_TIMEOUT) * time.Second,
		})
		if e != nil {
			return errors.Wrap(e, `create socks5 dialer fail`)
		}

		this.Conn, e = dialer.Dial("tcp", wa_host+":"+Sock_Port)

		return e
	} else { // no proxy
		this.Conn, e = net.DialTimeout(
			"tcp",
			wa_host+":"+Sock_Port,
			time.Duration(def.NET_TIMEOUT)*time.Second)
		return e
	}
}

func (this *Socket) Close() error {
	var e error

	if this.Conn != nil {
		e = this.Conn.Close()
	}

	this.wg.Wait()
	return e
}

// only return when all bytes are written
func (this *Socket) write_n(data []byte) error {
	offset := 0
	for offset < len(data) {
		nWrite, e := this.Conn.Write(data[offset:])
		if e != nil {
			return e
		}
		offset += nWrite
	}
	return nil
}
func (this *Socket) EnableReadTimeout(enable bool) {
	if enable {
		this.Conn.SetReadDeadline(time.Now().Add(time.Duration(def.NET_TIMEOUT) * time.Second))
	} else {
		this.Conn.SetReadDeadline(time.Time{})
	}
}
func (this *Socket) EnableWriteTimeout(enable bool) {
	if enable {
		this.Conn.SetWriteDeadline(time.Now().Add(time.Duration(def.NET_TIMEOUT) * time.Second))
	} else {
		this.Conn.SetWriteDeadline(time.Time{})
	}
}
func (this *Socket) WritePacket(data []byte) error {
	this.wLock.Lock()
	defer this.wLock.Unlock()

	this.wg.Add(1)
	defer this.wg.Done()

	if this.Conn == nil {
		return errors.New(`socket not connected`)
	}
	this.EnableWriteTimeout(true)
	defer this.EnableWriteTimeout(false)

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:], uint32(len(data)))

	pkt := append(buf[1:], data...) // 3-byte-len + data

	return this.write_n(pkt)
}

func (this *Socket) read_3_byte_len() (int, error) {
	bf, e := this.read_n(3)
	if e != nil {
		return 0, e
	}
	l := binary.BigEndian.Uint32(append([]byte{0}, bf[:]...))

	return int(l), nil
}
func (this *Socket) read_n(n int) ([]byte, error) {
	left := n
	ret := []byte{}

	for left > 0 {
		chunk := make([]byte, left)
		nRead, e := this.Conn.Read(chunk)
		if e != nil {
			return nil, e
		}
		left -= nRead
		ret = append(ret, chunk[0:nRead]...)
	}

	return ret, nil
}
func (this *Socket) ReadPacket() ([]byte, error) {
	this.rLock.Lock()
	defer this.rLock.Unlock()
	if this.Conn == nil {
		return nil, errors.New(`socket not connected`)
	}
	this.wg.Add(1)
	defer this.wg.Done()

	// read 3 byte
	body_len, e := this.read_3_byte_len()
	if e != nil {
		return nil, e
	}

	// read body
	return this.read_n(body_len)
}
