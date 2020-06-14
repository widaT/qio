package qio

import (
	"io"
)

type Buf struct {
	b []byte

	r int
	w int
}

func NewBuf() *Buf {
	return &Buf{
		b: make([]byte, 0x10000),
	}
}

func (buf *Buf) Read(b []byte) (n int, err error) {
	n = len(b)
	if n == 0 {
		return 0, nil
	}

	if buf.w == buf.r {
		return 0, io.EOF
	}

	n = min(n, buf.w-buf.r)
	copy(b, buf.b[buf.r:buf.r+n])
	buf.r += n
	return
}

func (buf *Buf) Buffered() (b []byte) {
	b = buf.b[buf.r:buf.w]
	buf.r = buf.w
	return
}

func (buf *Buf) Write(b []byte) (int, error) {
	n := copy(buf.b[buf.w:], b)
	buf.w += n
	//fmt.Println(buf.w)
	return n, nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
