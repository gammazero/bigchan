/*
Package bigchan provides a channel that uses a buffer between input and output.
The buffer can have any capacity specified.
*/
package bigchan

import "github.com/gammazero/queue"

// If capacity is this size or smaller, use a normal channel.
const normChanLimit = 16

// BigChan uses a queue to buffer data between the input and the output.
type BigChan struct {
	input, output chan interface{}
	length        chan int
	buffer        *queue.Queue
	capacity      int
}

// New creates a new BigChan with the specified buffer capacity.
//
// A capacity < 0 specifies unlimited capacity.  Use caution if specifying an
// unlimited capacity since no amount of storage is truly unlimited.
//
// If a capacity <= normChanLimit is given, then use a normal channel.
func New(capacity int) *BigChan {
	if capacity < 0 {
		capacity = -1
	} else if capacity <= normChanLimit {
		// Use normal channel
		ioChan := make(chan interface{}, capacity)
		ch := &BigChan{
			input:    ioChan,
			output:   ioChan,
			capacity: capacity,
		}
		return ch
	}

	ch := &BigChan{
		input:    make(chan interface{}),
		output:   make(chan interface{}),
		length:   make(chan int),
		buffer:   queue.New(0),
		capacity: capacity,
	}
	go ch.bufferInput()
	return ch
}

// In returns the write side of the channel.
func (ch *BigChan) In() chan<- interface{} {
	return ch.input
}

// Out returns the read side of the channel.
func (ch *BigChan) Out() <-chan interface{} {
	return ch.output
}

// Len returns the number of items buffered in the channel.
func (ch *BigChan) Len() int {
	if ch.length == nil {
		return len(ch.input)
	}
	return <-ch.length
}

// Cap returns the capacity of the channel.
func (ch *BigChan) Cap() int {
	return ch.capacity
}

// Close closes the channel.  Additional input will panic, output will continue
// to be readable until nil.
func (ch *BigChan) Close() {
	close(ch.input)
}

func (ch *BigChan) bufferInput() {
	var input, output, inputChan chan interface{}
	var next interface{}
	inputChan = ch.input
	input = inputChan

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				ch.buffer.Push(elem)
			} else {
				input = nil
			}
		case output <- next:
			ch.buffer.Pop()
		case ch.length <- ch.buffer.Length():
		}

		if ch.buffer.Length() > 0 {
			output = ch.output
			next = ch.buffer.Head()
			// If buffer at capacity, then stop accepting input.
			if ch.capacity != -1 && ch.buffer.Length() >= ch.capacity {
				input = nil
			} else {
				input = ch.input
			}
		} else {
			output = nil
			next = nil
		}
	}

	close(ch.output)
	close(ch.length)
}
