package qio

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
	length := len(b)
	if length == 0 || buf.w == buf.r {
		return 0, nil
	}
	movpos := min(length, buf.w-buf.r)
	copy(b, buf.b[buf.r:buf.r+movpos])
	buf.r += movpos
	n = movpos
	return
}

func (buf *Buf) Write(b []byte) (int, error) {
	buf.w += copy(buf.b[buf.w:], b)
	return len(b), nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
