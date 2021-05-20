package testhelpers

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
	Done       blockChan
	Mr         *MockedReader
	Mwc        *MockedWriteCloser
	FuncReader chanFuncReader
	FuncWriter chanFuncWriter
}

func (m *MockedConnection) NextReader() (int, io.Reader, error) {
	select {
	case <-m.Done:
		return 0, nil, websocket.ErrCloseSent
	case d := <-m.FuncReader:
		return d()
	}
}

func (m *MockedConnection) NextWriter(messageType int) (io.WriteCloser, error) {
	select {
	case <-m.Done:
		return nil, websocket.ErrCloseSent
	case d := <-m.FuncWriter:
		return d(messageType)
	}
}

func (m *MockedConnection) Close() error {
	select {
	case <-m.Done:
	default:
		close(m.Done)
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
		FuncReader: make(chanFuncReader, 5),
		FuncWriter: make(chanFuncWriter, 5),
		Mr:         pmr,
		Mwc:        pmwc,
		Done:       done,
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
