/*
Package bigchan implements a channel that uses a buffer between input and
output.  The buffer can have any capacity specified, from 0 (unbuffered) to
infinite.  This provides channel functionality with any buffer capacity
desired, and without taking up large amounts of storage when little is used.

If the specified buffer capacity is small, or unbuffered, the implementation is
provided by a normal chan.  When specifying an unlimited buffer capacity use
caution as the buffer is still limited by the resources available on the host
system.

Caution

The behavior of bigchan differs from the behavior of a normal channel in one
important way: After writing to the In() channel, the data may not be
immediately available on the Out() channel (until the buffer goroutine is
scheduled), and may be missed by a non-blocking select.

The ReadAny() function provides a way to read data is any is available.
Alternatively, the following is an example of how to handle this in your own
select block:

  var item interface{}
  var open bool
  select {
  case item, open = <-bigch.Out():
  default:
      // If using normal chan or buffer is empty.
      if bigch.Len() == 0 {
          // Return no data ready and chan open.
          return nil, true
      }
      item, open = <-ch.Out()
  }
  return item, open

Credits

This implementation is based on ideas/examples from:
https://github.com/eapache/channels

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
// unlimited capacity since storage is still limited by system resources.
//
// If a capacity <= normChanLimit is given, then use a normal channel.
func New(capacity int) *BigChan {
	if capacity < 0 {
		capacity = -1
	} else if capacity <= normChanLimit {
		// Use normal channel
		ioChan := make(chan interface{}, capacity)
		return &BigChan{
			input:    ioChan,
			output:   ioChan,
			capacity: capacity,
		}
	}

	ch := &BigChan{
		input:    make(chan interface{}),
		output:   make(chan interface{}),
		length:   make(chan int),
		buffer:   queue.New(),
		capacity: capacity,
	}
	go ch.bufferInput()
	return ch
}

// In returns the write side of the channel.
//
// Caution: After writing to the In() channel, the data may not be immediately
// available on the Out() channel, and may be missed by a non-blocking select.
// See ReadAny() for example solution.
func (ch *BigChan) In() chan<- interface{} {
	return ch.input
}

// Out returns the read side of the channel.
//
// Caution: After writing to the In() channel, the data may not be immediately
// available on the Out() channel, and may be missed by a non-blocking select.
// See ReadAny() for example solution.
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

// ReadAny reads an item if any is ready for reading, returning the item and a
// flag to indicate whether the channel is still open.
//
// This is useful for ensuring that a read from the bigchan will return data
// following a write to the bigchan.  This method also serves as an example of
// how to implement the same solution in select blocks outside of this code.
func (ch *BigChan) ReadAny() (item interface{}, open bool) {
	select {
	case item, open = <-ch.Out():
	default:
		// If using normal chan or buffer is empty.
		if ch.length == nil || <-ch.length == 0 {
			// Return no data ready and chan open.
			open = true
			return
		}
		item, open = <-ch.Out()
	}
	return
}

// bufferInput is the goroutine that transfers data from the In() chan to the
// buffer and from the buffer to the Out() chan.
func (ch *BigChan) bufferInput() {
	var input, output, inputChan chan interface{}
	var next interface{}
	inputChan = ch.input
	input = inputChan

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				// Push data from input chan to buffer.
				ch.buffer.Add(elem)
			} else {
				// Input chan closed; do not select input chan.
				input = nil
				inputChan = nil
			}
		case output <- next:
			// Wrote buffered data to output chan.  Remove item from buffer.
			ch.buffer.Remove()
		case ch.length <- ch.buffer.Length():
		}

		// If there is any data in the buffer, try to write it to output chan.
		if ch.buffer.Length() > 0 {
			output = ch.output
			next = ch.buffer.Peek()
			if ch.capacity != -1 {
				// If buffer at capacity, then stop accepting input.
				if ch.buffer.Length() >= ch.capacity {
					input = nil
				} else {
					input = inputChan
				}
			}
		} else {
			// No buffered data; do not select output chan.
			output = nil
			next = nil
		}
	}

	close(ch.output)
	close(ch.length)
}
