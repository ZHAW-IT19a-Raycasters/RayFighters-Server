package client

import (
	"bytes"
	"io"
	"sync"

	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/protocol/server"
)

type MockedHubClientInteraction struct {
	funcAllPlayers       func() []server.Player
	funcPlayerGotHit     func() (bool, server.Player)
	funcPlayerKD         func() []server.PlayerKD
	funcUnregisterClient func(meta ClientMeta) error
}

func (m *MockedHubClientInteraction) GetAllPlayers() []server.Player {
	return m.funcAllPlayers()
}

func (m *MockedHubClientInteraction) PlayerGotHit() (bool, server.Player) {
	return m.funcPlayerGotHit()
}

func (m *MockedHubClientInteraction) PlayersKD() []server.PlayerKD {
	return m.funcPlayerKD()
}

func (m *MockedHubClientInteraction) UnregisterClient(meta ClientMeta) error {
	return m.funcUnregisterClient(meta)
}

func newMockedHubClientInteraction() MockedHubClientInteraction {
	return MockedHubClientInteraction{
		funcAllPlayers: func() []server.Player {
			return []server.Player{}
		},
		funcPlayerGotHit: func() (bool, server.Player) {
			return false, server.NewPlayer(0, "")
		},
		funcPlayerKD: func() []server.PlayerKD {
			return []server.PlayerKD{}
		},
		funcUnregisterClient: func(meta ClientMeta) error {
			return nil
		},
	}
}

type MockedConnection struct {
	lock           sync.Mutex
	mr             *MockedReader
	mwc            *MockedWriteCloser
	funcNextReader func() (int, io.Reader, error)
	funcNextWriter func(messageType int) (io.WriteCloser, error)
	funcCloser     func() error
}

func (m *MockedConnection) NextReader() (int, io.Reader, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.funcNextReader()
}

func (m *MockedConnection) NextWriter(messageType int) (io.WriteCloser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
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
		lock: sync.Mutex{},
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
