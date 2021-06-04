package testhelpers

import (
	"bytes"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

// blockChan is a simple blocking channel abstraction.
type blockChan chan struct{}

// chanFuncReader is the channel overwhich the function
// for the reader interface is send over.
type chanFuncReader chan func() (int, io.Reader, error)

// chanFuncWriter is the channel overwhich the function
// for the writer interface is send over.
type chanFuncWriter chan func(messageType int) (io.WriteCloser, error)

// MockedConnection is a mock of a websocket connection.
type MockedConnection struct {
	Done       blockChan
	Mr         *MockedReader
	Mwc        *MockedWriteCloser
	FuncReader chanFuncReader
	FuncWriter chanFuncWriter
}

// NextReader is a blocking call for waiting for a message.
func (m *MockedConnection) NextReader() (int, io.Reader, error) {
	select {
	case <-m.Done:
		return 0, nil, websocket.ErrCloseSent
	case d := <-m.FuncReader:
		return d()
	}
}

// NextReader is a blocking call for waiting for writing interface.
func (m *MockedConnection) NextWriter(messageType int) (io.WriteCloser, error) {
	select {
	case <-m.Done:
		return nil, websocket.ErrCloseSent
	case d := <-m.FuncWriter:
		return d(messageType)
	}
}

// Close will close the connection.
func (m *MockedConnection) Close() error {
	select {
	case <-m.Done:
	default:
		close(m.Done)
	}
	return nil
}

// NewMockedConnection will initialize the mock
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

// MockedWriteCloser is an abstraction over the writer closer interface
// required by the mock.
type MockedWriteCloser struct {
	lock   *sync.Mutex
	buffer *bytes.Buffer
	cl     func() error
}

// newMockedWriteCloser will create a new MockedWriteCloser
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

// Bytes is part of the io.WriterCloser interface
func (mwc *MockedWriteCloser) Bytes() []byte {
	mwc.lock.Lock()
	defer mwc.lock.Unlock()
	return mwc.buffer.Bytes()
}

// Write is part of the io.WriterCloser interface
func (mwc *MockedWriteCloser) Write(p []byte) (n int, err error) {
	mwc.lock.Lock()
	defer mwc.lock.Unlock()
	return mwc.buffer.Write(p)
}

// Close is part of the io.WriterCloser interface
func (mwc *MockedWriteCloser) Close() error {
	mwc.lock.Lock()
	defer mwc.lock.Unlock()
	return mwc.cl()
}

// MockedReader is a mock of the io.Reader interface
type MockedReader struct {
	*bytes.Buffer
}

// newMockedReader will create a new MockedReader
func newMockedReader() MockedReader {
	return MockedReader{
		new(bytes.Buffer),
	}
}
