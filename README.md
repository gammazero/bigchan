# bigchan
[![Build Status](https://travis-ci.org/gammazero/bigchan.svg)](https://travis-ci.org/gammazero/bigchan)
[![GoDoc](https://godoc.org/github.com/gammazero/bugchan?status.svg)](https://godoc.org/github.com/gammazero/bigchan)

BigChan implements a channel that uses a buffer between input and output.  The buffer can have any capacity specified, from 0 (unbuffered) to infinite.  This provides channel functionality with any buffer capacity desired, and without taking up large amounts of storage when little is used.

If the specified buffer capacity is small, or unbuffered, the implementation is
provided by a normal chan.  When specifying an unlimited buffer capacity use caution as the buffer is still limited by the resources available on the host system.

The BigChan buffer is supplied by a fast queue implementation, which auto-resizes according to the number of items buffered. For more information on the queue, see: https://github.com/gammazero/queue


