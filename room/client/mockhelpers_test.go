package client

import (
	"bytes"
	"io"
	"sync"

	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/room/helpers"
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

func NewMockedConnection() MockedConnection {
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

type blockChan chan struct{}

type chanFuncReader chan func() (int, io.Reader, error)

type chanFuncWriter chan func(messageType int) (io.WriteCloser, error)

func setupMock(mock *MockedConnection) (blockChan, chanFuncReader, chanFuncWriter) {

	done := make(blockChan)

	// make the connections blocking, so that we can run other
	// stuff while it waits
	funcTransferReader := make(chanFuncReader, 10)
	funcTransferWriter := make(chanFuncWriter, 10)

	mock.mwc.cl = func() error {
		close(done)
		return nil
	}

	mock.funcNextReader = func() (int, io.Reader, error) {
		select {
		case <-done:
			return 0, nil, websocket.ErrCloseSent
		case d := <-funcTransferReader:
			return d()
		}
	}

	mock.funcNextWriter = func(messageType int) (io.WriteCloser, error) {
		select {
		case <-done:
			return nil, websocket.ErrCloseSent
		case d := <-funcTransferWriter:
			return d(messageType)
		}
	}

	// override the closer, so that the channels don't get changed by it
	mock.funcCloser = func() error {
		select {
		case <-done:
		default:
			close(done)
		}
		return nil
	}
	return done, funcTransferReader, funcTransferWriter
}

func helpTimeout() (helpers.HubTimer, func(id uint8) helpers.ClientTimer) {
	ht := helpers.NewHubTimer()
	f := func(id uint8) helpers.ClientTimer {
		return ht.NewClientTimer(id)
	}
	return ht, f
}
