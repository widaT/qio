module github.com/widaT/qio

go 1.14

require (
	github.com/libp2p/go-reuseport v0.0.1
	github.com/widaT/http1 v0.0.0-20200627122705-5d250421bc5e // indirect
	github.com/widaT/linkedbuf v0.0.0-20200627005813-e9045bdb9996
	github.com/widaT/poller v0.0.0-20200618102045-955b90a020f2
	github.com/widaT/tls13 v0.0.0-20200624074259-be05d3a87a28 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae
)

replace github.com/widaT/tls13 v0.0.0-20200624044940-6bc3b8e90328 => /home/wida/gocode/net/tls13

replace github.com/widaT/http1 v0.0.0-20200627122705-5d250421bc5e => /home/wida/gocode/net/http1
