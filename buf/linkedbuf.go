package buf

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

const GcFrequency int = 6
const BLOCKSIZE int = 4096

const (
	RefCountAdd   = 0
	RefCountMinus = 1
)

var blockPool sync.Pool

type Block struct {
	refCount   int32
	data       []byte
	blockIndex int
	next       *Block
}

func (b *Block) String() string {
	return fmt.Sprintf("blockInex:%d,refCount:%d", b.blockIndex, b.refCount)
}

func (b *Block) reset(blockIndex int) {
	b.refCount = 0
	b.blockIndex = blockIndex
	b.next = nil
}

type Segment struct {
	block  *Block
	b      []byte
	droped bool
}

func (d *Segment) Byte() []byte {
	return d.b
}
func (d *Segment) Drop() {
	if d.droped {
		return
	}
	refCount(d.block, RefCountMinus)
}

func NewBlock(blockIndex int) *Block {
	b := blockPool.Get()
	if b != nil {
		block := b.(*Block)
		block.reset(blockIndex)
		return block
	}

	return &Block{
		data:       make([]byte, BLOCKSIZE),
		blockIndex: blockIndex,
	}
}

type Point struct {
	b   *Block
	pos int
}

type LinkedBuffer struct {
	l              *list.List
	nextBlockIndex int
	wp             Point
	rp             Point
}

func New() *LinkedBuffer {
	l := list.New()
	block := NewBlock(0)
	l.PushBack(block)
	return &LinkedBuffer{
		l: l,
		rp: Point{
			b:   block,
			pos: 0,
		},
		wp: Point{
			b:   block,
			pos: 0,
		},
		nextBlockIndex: 1,
	}
}

func (buf *LinkedBuffer) growth() {
	if buf.nextBlockIndex%GcFrequency == 0 {
		buf.Gc()
	}
	block := NewBlock(buf.nextBlockIndex)
	buf.wp.b.next = block
	buf.l.PushBack(block)
	buf.nextBlockIndex++
	buf.wp.pos = 0
	buf.wp.b = block
}

func (buf *LinkedBuffer) NexWriteBlock() []byte {
	if buf.wp.pos == BLOCKSIZE {
		buf.growth()
	}
	return buf.wp.b.data[buf.wp.pos:]
}

func (buf *LinkedBuffer) MoveWritePiont(n int) (s *Segment) {
	s = new(Segment)
	s.block = buf.wp.b
	s.b = buf.wp.b.data[buf.wp.pos : buf.wp.pos+n]
	buf.wp.pos += n
	buf.wp.b.refCount++
	return s
}

func (buf *LinkedBuffer) Bytes() []byte {
	n := buf.Buffered()
	wp := buf.wp
	rp := buf.rp
	left := BLOCKSIZE - rp.pos
	if n <= left {
		return rp.b.data[rp.pos:wp.pos]
	}
	b := make([]byte, n)
	nn := 0
	nn += copy(b, rp.b.data[rp.pos:])
	block := rp.b
	for block.next != nil {
		block = block.next
		nn += copy(b[nn:], block.data[:min(n-nn, BLOCKSIZE)])
	}
	return b
}

func (buf *LinkedBuffer) Shift(n int) {
	if n == 0 {
		return
	}
	rp := buf.rp
	left := BLOCKSIZE - rp.pos
	if n <= left {
		buf.rp.pos += n
		return
	}
	if n > buf.Buffered() {
		n = buf.Buffered()
	}
	nn := left
	block := rp.b
	pos := 0
	for block.next != nil {
		if nn >= n {
			break
		}
		block = block.next
		buf.rp.b = block
		pos = min(n-nn, BLOCKSIZE)
		nn += pos
		buf.rp.pos = pos
	}
}

func (buf *LinkedBuffer) Len() int {
	return buf.Buffered()
}

func (buf *LinkedBuffer) Read(b []byte) (n int, err error) {
	n = len(b)
	if n == 0 {
		return
	}
	wp := buf.wp
	rp := buf.rp
	if wp.b == rp.b && wp.pos == rp.pos {
		err = io.EOF
		return
	}
	if n > buf.Buffered() {
		n = buf.Buffered()
	}
	nn := 0
	if len(rp.b.data[rp.pos:]) >= n {
		buf.rp.pos += copy(b, rp.b.data[rp.pos:rp.pos+n])
		return
	}

	nn += copy(b, rp.b.data[rp.pos:])
	block := rp.b
	for block.next != nil {
		block = block.next
		buf.rp.b = block
		buf.rp.pos = copy(b[nn:], block.data[:min(n-nn, BLOCKSIZE)])
		nn += buf.rp.pos
	}
	return
}

func (buf *LinkedBuffer) Release() {
	var next *list.Element
	for item := buf.l.Front(); item != nil; item = next {
		next = item.Next()
		block := item.Value.(*Block)
		buf.l.Remove(item)
		blockPool.Put(block)
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (buf *LinkedBuffer) Buffered() int {
	wp := buf.wp
	rp := buf.rp
	n := wp.b.blockIndex - rp.b.blockIndex
	if n == 0 {
		return wp.pos - rp.pos
	}
	return (BLOCKSIZE - rp.pos + wp.pos) + (n-1)*BLOCKSIZE
}

/* func (buf *LinkedBuffer) Buffered() int {

	for p:= buf.wp;
} */

func (buf *LinkedBuffer) BlockLen() int {
	return buf.l.Len()
}

func (buf *LinkedBuffer) Range(fn func(*Block)) {
	for item := buf.l.Front(); item != nil; item = item.Next() {
		block := item.Value.(*Block)
		fn(block)
	}
}

func (buf *LinkedBuffer) Gc() {
	var next *list.Element
	for item := buf.l.Front(); item != nil; item = next {
		next = item.Next()
		block := item.Value.(*Block)
		if block == buf.rp.b {
			break
		}
		/* 		if atomic.LoadInt32(&block.refCount) == 0 {
		   			buf.l.Remove(item)
		   			blockPool.Put(block)
		   		} else {
		   			break
		   		} */
		buf.l.Remove(item)
		blockPool.Put(block)
	}
}

func refCount(b *Block, op int) {
	if op == RefCountAdd {
		atomic.AddInt32(&b.refCount, 1)
	} else {
		atomic.AddInt32(&b.refCount, -1)
	}
}

type ConpositeBuf struct {
	segments []*Segment
	length   int
	read     int
}

func (c *ConpositeBuf) Wrap(s *Segment) {
	c.segments = append(c.segments, s)
	c.length += len(s.b)
}

func (c *ConpositeBuf) Buffered() int {
	return c.length - c.read
}

func (c *ConpositeBuf) Read(b []byte) (n int, err error) {
	if c.length == 0 {
		return 0, errors.New("no data")
	}
	if c.read > c.length {
		return 0, io.EOF
	}
	count := len(b)
	if count == 0 {
		return 0, nil
	}

	//found read postion
	var j = 0
	var pos = 0
	var sIndex = 0
	for i, s := range c.segments {
		if j+len(s.b) > c.read {
			sIndex = i
			pos = c.read - j
			break
		} else {
			j += len(s.b)
		}
	}

	//copy data
	for ; sIndex < len(c.segments); sIndex++ {
		block := c.segments[sIndex].b[pos:]
		if n >= count-1 {
			break
		}
		num := copy(b[n:], block)
		c.read += num
		n += num
		pos = 0
	}
	return
}

func (c *ConpositeBuf) ReadN(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("Parameter error")
	}
	if c.length < n {
		return nil, errors.New("not enough data")
	}

	if n <= len(c.segments[0].b) {
		c.read += n
		return c.segments[0].b[:n], nil
	}
	bigB := make([]byte, n)
	c.Read(bigB)

	return bigB, nil
}

//Peek return bytes but not move the read point
func (c *ConpositeBuf) Peek(n int) (b []byte, err error) {
	if n > c.Buffered() {
		return nil, errors.New("not enough data")
	}
	b, err = c.ReadN(n)
	if err != nil {
		return
	}
	c.read -= len(b)
	return
}

func (c *ConpositeBuf) Discurd(n int) error {
	if n < 0 {
		return errors.New("Parameter error")
	}

	if n > c.length {
		return errors.New("not enough data")
	}

	if c.read+n <= c.length {
		c.read += n
	}
	return nil
}

func (c *ConpositeBuf) Drop() {
	for _, s := range c.segments {
		s.Drop()
	}
}
