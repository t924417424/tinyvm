package internal

import (
	"bytes"
	"sync"
)

type LXDExecIO struct {
	write  bytes.Buffer
	read   bytes.Buffer
	m      sync.Locker
	signal chan struct{}
	call   func([]byte) []byte
}

func NewExecIO(call func([]byte) []byte) *LXDExecIO {
	m := &sync.Mutex{}
	return &LXDExecIO{
		write:  bytes.Buffer{},
		read:   bytes.Buffer{},
		m:      m,
		signal: make(chan struct{}),
		call:   call,
	}
}

func (w *LXDExecIO) Write(p []byte) (int, error) {
	w.m.Lock()
	len, err := w.write.Write(p)
	defer w.m.Unlock()
	if err == nil {
		out := w.call(p)
		if out != nil {
			w.read.Write(out)
			w.signal <- struct{}{}
		}
	}
	return len, err
}

func (w *LXDExecIO) Close() error {
	w.m.Lock()
	defer w.m.Unlock()
	w.write.Reset()
	w.read.Reset()
	return nil
}

func (r *LXDExecIO) Read(p []byte) (n int, err error) {
	// r.m.Lock()
	// defer r.m.Unlock()
	<-r.signal
	len, err := r.read.Read(p)
	println(string(p))
	return len, err
}
