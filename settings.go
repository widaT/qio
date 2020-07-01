package qio

import(
	"runtime"
)

type Option func(*Settings)

type Settings struct{
	keepAlive       bool
	keepAlivePeriod int //second count use 3 means 3 second ,notice that it's different from conn.SetKeepAlivePeriod
	portReuse       bool
	eventLoopNum    int
}

func defaultSetting()*Settings{
	return &Settings{
		eventLoopNum : runtime.NumCPU(),
	}
}

func SetKeepAlive(secs int) Option {
	return func(s *Settings) {
		s.keepAlive = true
		s.keepAlivePeriod = secs
	}
}

func SetPortReuse(b bool) Option{
	return func(s *Settings) {
		s.portReuse = b
	}
}

func SetEventLoopNum(int int) Option{
	return func(s *Settings) {
		s.eventLoopNum = int
	}
}
