package bigchan

import (
	"testing"
	"time"
)

func TestBigChanUnlimited(t *testing.T) {
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

func TestBigChanLimited(t *testing.T) {
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

func TestBigChanLimit(t *testing.T) {
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

func TestBigChanRace(t *testing.T) {
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

func BenchmarkBigChanSerial(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkBigChanParallel(b *testing.B) {
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

func BenchmarkBigChanPushPull(b *testing.B) {
	ch := New(b.N)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
		<-ch.Out()
	}
}
