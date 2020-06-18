package qio

import "github.com/widaT/qio/conn"

type EventServer interface {
	OnContect(conn.Conn)
	OnMessage(conn.Conn) error
	OnClose(conn.Conn)
}

type DefaultEvServer struct{}

func (e *DefaultEvServer) OnContect(c conn.Conn) {
}
func (e *DefaultEvServer) OnMessage(c conn.Conn) error {
	return nil
}
func (e *DefaultEvServer) OnClose(c conn.Conn) {
}
