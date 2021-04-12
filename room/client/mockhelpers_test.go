package client

import (
	"bytes"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

type MockedConnection struct {
	lock           sync.RWMutex
	mr             *MockedReader
	mwc            *MockedWriteCloser
	funcNextReader func() (int, io.Reader, error)
	funcNextWriter func(messageType int) (io.WriteCloser, error)
	funcCloser     func() error
}

func (m *MockedConnection) NextReader() (int, io.Reader, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.funcNextReader()
}

func (m *MockedConnection) NextWriter(messageType int) (io.WriteCloser, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.funcNextWriter(messageType)
}

func (m *MockedConnection) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	// replace the next func reader and next writer with error returning versions
	m.funcNextReader = func() (int, io.Reader, error) {
		return 0, nil, websocket.ErrCloseSent
	}
	m.funcNextWriter = func(messageType int) (io.WriteCloser, error) {
		return nil, websocket.ErrCloseSent
	}
	return m.funcCloser()
}

func newMockedConnection() MockedConnection {
	mr := newMockedReader()
	mwc := newMockedWriteCloser()
	pmr := &mr
	pmwc := &mwc
	return MockedConnection{
		lock: sync.RWMutex{},
		mr:   pmr,
		mwc:  pmwc,
		funcNextReader: func() (int, io.Reader, error) {
			return 0, nil, websocket.ErrCloseSent
		},
		funcNextWriter: func(messageType int) (io.WriteCloser, error) {
			return nil, websocket.ErrCloseSent
		},
		funcCloser: func() error {
			return nil
		},
	}
}

type MockedWriteCloser struct {
	*bytes.Buffer
	cl func() error
}

func newMockedWriteCloser() MockedWriteCloser {
	bff := new(bytes.Buffer)
	cl := func() error { return nil }
	return MockedWriteCloser{
		bff,
		cl,
	}
}

func (mwc *MockedWriteCloser) Close() error {
	return mwc.cl()
}

type MockedReader struct {
	*bytes.Buffer
}

func newMockedReader() MockedReader {
	return MockedReader{
		new(bytes.Buffer),
	}
}
