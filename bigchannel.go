package bigchannel

import "github.com/gammazero/queue"

// If capacity is this size of smaller, use a normal channel.
const normChanLimit = 16

// BigChannel uses a queue to buffer data between the input and the output.
type BigChannel struct {
	input, output chan interface{}
	length        chan int
	buffer        *queue.Queue
	capacity      int
}

func New(capacity int) *BigChannel {
	if capacity < 0 {
		capacity = -1
	} else if capacity <= normChanLimit {
		// Use normal channel
		ioChan := make(chan interface{}, capacity)
		ch := &BigChannel{
			input:    ioChan,
			output:   ioChan,
			capacity: capacity,
		}
		return ch
	}

	ch := &BigChannel{
		input:    make(chan interface{}),
		output:   make(chan interface{}),
		length:   make(chan int),
		buffer:   queue.New(0),
		capacity: capacity,
	}
	go ch.bufferInput()
	return ch
}

func (ch *BigChannel) In() chan<- interface{} {
	return ch.input
}

func (ch *BigChannel) Out() <-chan interface{} {
	return ch.output
}

func (ch *BigChannel) Len() int {
	if ch.length == nil {
		return len(ch.input)
	}
	return <-ch.length
}

func (ch *BigChannel) Cap() int {
	return ch.capacity
}

func (ch *BigChannel) Close() {
	close(ch.input)
}

func (ch *BigChannel) bufferInput() {
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
