package optipool

import (
	"errors"
	"testing"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
		bufSize  int
		wantSize int
		wantBuf  int
		wantErr  bool
	}{
		{"valid pool", 10, 1024, 10, 1024, false},
		{"single buffer", 1, 512, 1, 512, false},
		{"large pool", 100000, 256, 100000, 256, false},
		{"zero pool size", 0, 1024, 0, 1024, false},
		{"zero buf size", 10, 0, 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool(tt.poolSize, tt.bufSize)
			if p == nil {
				t.Fatal("NewPool returned nil")
			}

			if cap(p.pool) != tt.wantSize {
				t.Errorf("pool capacity = %d, want %d", cap(p.pool), tt.wantSize)
			}
			if p.bufSize != tt.wantBuf {
				t.Errorf("bufSize = %d, want %d", p.bufSize, tt.wantBuf)
			}

			for i := 0; i < tt.poolSize; i++ {
				buf := <-p.pool
				if cap(buf) != tt.bufSize {
					t.Errorf("buffer %d capacity = %d, want %d", i, cap(buf), tt.bufSize)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := NewPool(5, 1024)
		buf, err := p.Get()
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if buf == nil {
			t.Fatal("Get() returned nil buffer")
		}
		if cap(buf) != 1024 {
			t.Errorf("buffer capacity = %d, want 1024", cap(buf))
		}
	})

	t.Run("empty pool", func(t *testing.T) {
		p := NewPool(0, 1024)
		_, err := p.Get()
		if err == nil {
			t.Fatal("Get() expected ErrPoolEmpty, got nil")
		}
		if !errors.Is(err, ErrPoolEmpty) {
			t.Errorf("Get() error = %v, want ErrPoolEmpty", err)
		}
	})

	t.Run("pool exhausted", func(t *testing.T) {
		p := NewPool(2, 256)
		_, _ = p.Get()
		_, _ = p.Get()
		_, err := p.Get()
		if !errors.Is(err, ErrPoolEmpty) {
			t.Errorf("Get() error = %v, want ErrPoolEmpty", err)
		}
	})
}

func TestPut(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := NewPool(5, 1024)
		buf, _ := p.Get()
		buf2 := make([]byte, 1024)
		err := p.Put(buf2)
		if err != nil {
			t.Fatalf("Put() error = %v", err)
		}
		_ = p.Put(buf)
	})

	t.Run("nil buffer", func(t *testing.T) {
		p := NewPool(5, 1024)
		err := p.Put(nil)
		if err == nil {
			t.Fatal("Put(nil) expected ErrAlienBuffer, got nil")
		}
		if !errors.Is(err, ErrAlienBuffer) {
			t.Errorf("Put(nil) error = %v, want ErrAlienBuffer", err)
		}
	})

	t.Run("wrong capacity", func(t *testing.T) {
		p := NewPool(5, 1024)
		buf := make([]byte, 512)
		err := p.Put(buf)
		if err == nil {
			t.Fatal("Put(wrong cap) expected ErrAlienBuffer, got nil")
		}
		if !errors.Is(err, ErrAlienBuffer) {
			t.Errorf("Put(wrong cap) error = %v, want ErrAlienBuffer", err)
		}
	})

	t.Run("full pool", func(t *testing.T) {
		p := NewPool(1, 1024)
		buf1, _ := p.Get()
		_ = p.Put(buf1)
		buf2 := make([]byte, 1024)
		err := p.Put(buf2)
		if err == nil {
			t.Fatal("Put() on full pool expected ErrPoolFull, got nil")
		}
		if !errors.Is(err, ErrPoolFull) {
			t.Errorf("Put() on full pool error = %v, want ErrPoolFull", err)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		p := NewPool(1, 256)
		buf, _ := p.Get()
		buf[0] = 42
		err := p.Put(buf)
		if err != nil {
			t.Fatalf("Put() error = %v", err)
		}
		buf2, _ := p.Get()
		if buf2[0] != 42 {
			t.Errorf("buffer data not preserved: got %d, want 42", buf2[0])
		}
	})
}

func TestPoolConcurrency(t *testing.T) {
	p := NewPool(100, 1024)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				buf, err := p.Get()
				if err != nil {
					continue
				}
				buf[0] = byte(j)
				_ = p.Put(buf)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
