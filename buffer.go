// Go MySQL Driver - A MySQL-Driver for Go's database/sql package
//
// Copyright 2013 The Go-MySQL-Driver Authors. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package mysql

import "io"

const defaultBufSize = 4096

// A buffer which is used for both reading and writing.
// This is possible since communication on each connection is synchronous.
// In other words, we can't write and read simultaneously on the same connection.
// The buffer is similar to bufio.Reader / Writer but zero-copy-ish
// Also highly optimized for this particular use case.
type buffer struct {
	buf    []byte
	rd     io.Reader
	idx    int
	length int
}

func newBuffer(rd io.Reader) *buffer {
	var b [defaultBufSize]byte
	return &buffer{
		buf: b[:],
		rd:  rd,
	}
}

// fill reads into the buffer until at least _need_ bytes are in it
func (b *buffer) fill(need int) (err error) {
	// move existing data to the beginning
	if b.length > 0 && b.idx > 0 {
		copy(b.buf[0:b.length], b.buf[b.idx:])
	}

	// grow buffer if necessary
	// TODO: let the buffer shrink again at some point
	//       Maybe keep the org buf slice and swap back?
	if need > len(b.buf) {
		// Round up to the next multiple of the default size
		newBuf := make([]byte, ((need/defaultBufSize)+1)*defaultBufSize)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}

	b.idx = 0

	var n int
	for {
		n, err = b.rd.Read(b.buf[b.length:])
		b.length += n

		if b.length < need && err == nil {
			continue
		}
		return // err
	}
	return
}

// returns next N bytes from buffer.
// The returned slice is only guaranteed to be valid until the next read
func (b *buffer) readNext(need int) (p []byte, err error) {
	if b.length < need {
		// refill
		err = b.fill(need) // err deferred
		if err == io.EOF && b.length >= need {
			err = nil
		}
	}

	p = b.buf[b.idx : b.idx+need]
	b.idx += need
	b.length -= need
	return
}

// returns a buffer with the requested size.
// If possible, a slice from the existing buffer is returned.
// Otherwise a bigger buffer is made.
// Only one buffer (total) can be used at a time.
func (b *buffer) takeBuffer(length int) []byte {
	if b.length > 0 {
		return nil
	}

	// test (cheap) general case first
	if length <= defaultBufSize || length <= cap(b.buf) {
		return b.buf[:length]
	}

	if length < maxPacketSize {
		b.buf = make([]byte, length)
		return b.buf
	}
	return make([]byte, length)
}

// shortcut which can be used if the requested buffer is guaranteed to be
// smaller than defaultBufSize
// Only one buffer (total) can be used at a time.
func (b *buffer) takeSmallBuffer(length int) []byte {
	if b.length == 0 {
		return b.buf[:length]
	}
	return nil
}

// takeCompleteBuffer returns the complete existing buffer.
// This can be used if the necessary buffer size is unknown.
// Only one buffer (total) can be used at a time.
func (b *buffer) takeCompleteBuffer() []byte {
	if b.length == 0 {
		return b.buf
	}
	return nil
}

var fieldCache = make(chan []mysqlField, 16)

func makeFields(n int) []mysqlField {
	select {
	case f := <-fieldCache:
		if cap(f) >= n {
			return f[:n]
		}
	default:
	}
	return make([]mysqlField, n)
}

func putFields(f []mysqlField) {
	select {
	case fieldCache <- f:
	default:
	}
}

var rowsCache = make(chan *mysqlRows, 16)

func newMysqlRows() *mysqlRows {
	select {
	case r := <-rowsCache:
		return r
	default:
		return new(mysqlRows)
	}
}

func putMysqlRows(r *mysqlRows) {
	*r = mysqlRows{} // zero it
	select {
	case rowsCache <- r:
	default:
	}
}
