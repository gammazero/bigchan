package bigchan

import (
	"testing"
	"time"
)

func TestUnlimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := New(-1)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < msgCount; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func TestLimitedSpace(t *testing.T) {
	const msgCount = 1000
	ch := New(normChanLimit * 2)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < msgCount; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func TestBufferLimit(t *testing.T) {
	ch := New(normChanLimit * 2)
	for i := 0; i < normChanLimit*2; i++ {
		ch.In() <- nil
	}
	var timeout bool
	select {
	case ch.In() <- nil:
	case <-time.After(200 * time.Millisecond):
		timeout = true
	}
	if !timeout {
		t.Fatal("expected timeout on full channel")
	}
}

func TestReadAny(t *testing.T) {
	const msgCount = 1000000
	ch := New(-1)
	for i := 0; i < msgCount; i++ {
		ch.In() <- i
		o, _ := ch.ReadAny()
		if o.(int) != i {
			t.Fatal("Fail to read channel after it was written.")
		}
	}

	// This is equivalent to above.
	var o interface{}
	for i := 0; i < msgCount; i++ {
		ch.In() <- i
		select {
		case o, _ = <-ch.Out():
		default:
			if ch.Len() == 0 {
				t.Fatal("Fail to read channel after it was written.")
			}
			o, _ = <-ch.Out()
		}
		if o.(int) != i {
			t.Fatal("Fail to read channel after it was written.")
		}
	}
}

func TestRace(t *testing.T) {
	ch := New(-1)
	go ch.Len()
	go ch.Cap()

	go func() {
		ch.In() <- nil
	}()

	go func() {
		<-ch.Out()
	}()
}

func TestDouble(t *testing.T) {
	const msgCount = 1000
	ch := New(100)
	recvCh := New(100)
	go func() {
		for i := 0; i < msgCount; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	go func() {
		for i := 0; i < msgCount; i++ {
			val := <-ch.Out()
			if i != val.(int) {
				t.Fatal("expected", i, "but got", val.(int))
			}
			recvCh.In() <- i
		}
	}()
	for i := 0; i < msgCount; i++ {
		val := <-recvCh.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func BenchmarkSerial(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkParallel(b *testing.B) {
	ch := New(b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch.Out()
		}
		<-ch.Out()
	}()
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	ch.Close()
}

func BenchmarkPushPull(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
		<-ch.Out()
	}
}

func BenchmarkReadAny(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
		ch.ReadAny()
	}
}
