package rpc

import (
	"sync"

	"wa/rpc/pb"
)

type Stream struct {
	stream  pb.Rpc_PushServer
	chClose chan struct{}

	mu      sync.RWMutex
	isAlive bool
}

func NewStream(con pb.Rpc_PushServer) *Stream {
	return &Stream{
		stream:  con,
		chClose: make(chan struct{}),
	}
}
func (st *Stream) Close() {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if st.isAlive {
		st.chClose <- struct{}{}
	}
}
func (st *Stream) SetAlive(alive bool) {
	st.mu.Lock()
	st.isAlive = alive
	st.mu.Unlock()
}

func (st *Stream) KeepAlive() {
	st.SetAlive(true)

	defer st.SetAlive(false)

	ctx := st.stream.Context()
	select {
	case <-st.chClose:
		break
	case <-ctx.Done():
		break
	}
}
