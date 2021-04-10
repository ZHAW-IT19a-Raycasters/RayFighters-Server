package client

import "io"

// ClientConenction is an iterface that is implemented by the
// websocket.Conn struct, if all
type ClientConnection interface {
	NextReader() (int, io.Reader, error)
	NextWriter(messageType int) (io.WriteCloser, error)
	Close() error
}
