package optipool

import "errors"

type Pool struct {
	pool    chan []byte
	bufSize int
}

var (
	ErrAlienBuffer = errors.New("Buffer capacity does not match the pool it is Put into.")
	ErrPoolFull    = errors.New("Pool is full.")
	ErrPoolEmpty   = errors.New("Pool is empty.")
)

// Creates a new pool with poolSize buffers of bufSize capacity.
func NewPool(poolSize int, bufSize int) *Pool {
	p := make(chan []byte, poolSize)
	for range poolSize {
		buf := make([]byte, bufSize)
		p <- buf
	}
	return &Pool{
		pool:    p,
		bufSize: bufSize,
	}
}

// Gets a buffer from the pool. If pool gets empty during runtime - create a bigger pool.
func (p *Pool) Get() ([]byte, error) {
	select {
	case buf := <-p.pool:
		return buf, nil
	default:
		return nil, ErrPoolEmpty
	}
}

// Puts a buffer back into the pool. If pool gets full during runtime - there is an ownership issue in application code.
func (p *Pool) Put(buf []byte) error {
	if buf == nil || cap(buf) != p.bufSize {
		return ErrAlienBuffer
	}
	select {
	case p.pool <- buf:
		return nil
	default:
		return ErrPoolFull
	}
}
