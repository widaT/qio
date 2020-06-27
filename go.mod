module github.com/widaT/qio

go 1.14

require (
	github.com/widaT/http1 v0.0.0-20200623114559-de092b2f1000 // indirect
	github.com/widaT/linkedbuf v0.0.0-20200627005813-e9045bdb9996
	github.com/widaT/poller v0.0.0-20200618102045-955b90a020f2
	github.com/widaT/tls13 v0.0.0-20200624044940-6bc3b8e90328 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae
)

replace github.com/widaT/tls13 v0.0.0-20200624044940-6bc3b8e90328 => /home/wida/gocode/net/tls13

replace github.com/widaT/http1 v0.0.0-20200623114559-de092b2f1000 => /home/wida/gocode/net/http1
