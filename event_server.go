package qio

type EventServer interface {
	OnConnect(*Conn) error
	OnMessage(*Conn) error
	OnClose(*Conn)
}

type DefaultEvServer struct{}

func (e *DefaultEvServer) OnConnect(c *Conn) error {
	return nil
}
func (e *DefaultEvServer) OnMessage(c *Conn) error {
	return nil
}
func (e *DefaultEvServer) OnClose(c *Conn) {
}
