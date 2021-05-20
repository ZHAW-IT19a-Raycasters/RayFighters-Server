package controller

import (
	"bytes"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

type blockChan chan struct{}

type chanFuncReader chan func() (int, io.Reader, error)

type chanFuncWriter chan func(messageType int) (io.WriteCloser, error)

type MockedConnection struct {
	done       blockChan
	mr         *MockedReader
	mwc        *MockedWriteCloser
	funcReader chanFuncReader
	funcWriter chanFuncWriter
}

func (m *MockedConnection) NextReader() (int, io.Reader, error) {
	select {
	case <-m.done:
		return 0, nil, websocket.ErrCloseSent
	case d := <-m.funcReader:
		return d()
	}
}

func (m *MockedConnection) NextWriter(messageType int) (io.WriteCloser, error) {
	select {
	case <-m.done:
		return nil, websocket.ErrCloseSent
	case d := <-m.funcWriter:
		return d(messageType)
	}
}

func (m *MockedConnection) Close() error {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	return nil
}

func NewMockedConnection() MockedConnection {
	mr := newMockedReader()
	mwc := newMockedWriteCloser()
	pmr := &mr
	pmwc := &mwc
	done := make(blockChan)

	return MockedConnection{
		mr:   pmr,
		mwc:  pmwc,
		done: done,
	}
}

type MockedWriteCloser struct {
	lock   *sync.Mutex
	buffer *bytes.Buffer
	cl     func() error
}

func newMockedWriteCloser() MockedWriteCloser {
	bff := new(bytes.Buffer)
	cl := func() error { return nil }
	lock := &sync.Mutex{}
	return MockedWriteCloser{
		lock,
		bff,
		cl,
	}
}

func (mwc *MockedWriteCloser) Bytes() []byte {
	mwc.lock.Lock()
	defer mwc.lock.Unlock()
	return mwc.buffer.Bytes()
}

func (mwc *MockedWriteCloser) Write(p []byte) (n int, err error) {
	mwc.lock.Lock()
	defer mwc.lock.Unlock()
	return mwc.buffer.Write(p)
}

func (mwc *MockedWriteCloser) Close() error {
	mwc.lock.Lock()
	defer mwc.lock.Unlock()
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
