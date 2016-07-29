package bigchan

import (
	"testing"
	"time"
)

func TestBigChan(t *testing.T) {
	ch := New(-1)
	go func() {
		for i := 0; i < 1000; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < 1000; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("expected", i, "but got", val.(int))
		}
	}
}

func TestBigChanLimit(t *testing.T) {
	ch := New(20)
	for i := 0; i < 20; i++ {
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

func BenchmarkBigChanSerial(b *testing.B) {
	ch := New(-1)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkBigChanParallel(b *testing.B) {
	ch := New(-1)
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
	ch := New(-1)
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
		<-ch.Out()
	}
}
