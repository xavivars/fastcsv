package csv

import (
	"fmt"
	"io"
	"unicode/utf8"
)

type bufferedReader struct {
	r      io.Reader
	data   []byte
	cursor int
}

func (b *bufferedReader) peekRune() (rune, int, error) {
START:
	r, size := utf8.DecodeRune(b.data[b.cursor:])
	if r == utf8.RuneError {
		if size == 0 {
			if err := b.more(); err != nil {
				return r, size, err
			}
			goto START
		}
		if size == 1 {
			return r, size, fmt.Errorf("Invalid UTF-8")
		}
	}
	return r, size, nil
}

func (b *bufferedReader) nextRune() (rune, int, error) {
	r, size, err := b.peekRune()
	b.cursor += size
	return r, size, err
}

func (b *bufferedReader) more() error {
	if len(b.data) == cap(b.data) {
		temp := make([]byte, len(b.data), 2*len(b.data)+1)
		copy(temp, b.data)
		b.data = temp
	}

	// read the new bytes onto the end of the buffer
	n, err := b.r.Read(b.data[len(b.data):cap(b.data)])
	b.data = b.data[:len(b.data)+n]
	return err
}

func (b *bufferedReader) reset() {
	copy(b.data, b.data[b.cursor:])
	b.data = b.data[:len(b.data)-b.cursor]
	b.cursor = 0
}

func (b *bufferedReader) slice(start, end int) []byte {
	return b.data[start:end]
}

func (b *bufferedReader) sliceFrom(start int) []byte {
	return b.data[start:b.cursor]
}

type fields struct {
	fieldStart int
	buffer     bufferedReader
	hitEOL     bool
	field      []byte
	err        error
}

func (fs *fields) Reset() {
	fs.buffer.reset()
	fs.field = nil
	fs.fieldStart = 0
	fs.hitEOL = false
}

func (fs *fields) nextUnquotedField() bool {
	const sizeEOL = 1
	const sizeDelim = 1
	for {
		ch, _, err := fs.buffer.nextRune()
		if err != nil {
			if err == io.EOF {
				fs.field = fs.buffer.data[fs.fieldStart:fs.buffer.cursor]
				fs.hitEOL = true
				fs.err = err
				return true
			}
			fs.err = err
			return false
		}

		switch ch {
		case ',':
			fs.field = fs.buffer.data[fs.fieldStart : fs.buffer.cursor-sizeDelim]
			fs.fieldStart = fs.buffer.cursor
			return true
		case '\n':
			fs.field = fs.buffer.data[fs.fieldStart : fs.buffer.cursor-sizeEOL]
			fs.hitEOL = true
			return true
		default:
			continue
		}
	}
}

func nextQuotedField(buffer *bufferedReader) ([]byte, bool, error) {
	// skip past the initial quote rune; we already checked the error when we
	// peeked it before this method call, so no need to handle it again
	buffer.nextRune()
	start := buffer.cursor

	writeCursor := buffer.cursor
	last := rune(0)
	for {
		r, size, err := buffer.nextRune()
		if err != nil {
			return buffer.data[start:writeCursor], true, err
		}
		switch r {
		case ',':
			if last == '"' {
				return buffer.data[start:writeCursor], false, nil
			}
		case '\n':
			if last == '"' {
				return buffer.data[start:writeCursor], true, nil
			}
		case '"':
			if last != '"' { // skip the first '"'
				last = r
				continue
			}
		}
		writeCursor += size
		last = r
		// copy the current rune onto writeCursor if writeCursor !=
		// buffer.cursor
		if writeCursor != buffer.cursor {
			copy(
				buffer.data[writeCursor:writeCursor+size],
				buffer.data[buffer.cursor:buffer.cursor+size],
			)
		}
	}
}

func (fs *fields) next() bool {
	if fs.hitEOL {
		return false
	}
	first, _, err := fs.buffer.peekRune()
	if err != nil {
		fs.err = err
		return false
	}

	if first == '"' {
		fs.field, fs.hitEOL, fs.err = nextQuotedField(&fs.buffer)
		return fs.err == nil || fs.err == io.EOF
	}
	return fs.nextUnquotedField()
}

type Reader struct {
	fields       fields
	fieldsBuffer [][]byte
}

func (r *Reader) Read() ([][]byte, error) {
	if err := r.fields.err; err != nil {
		return nil, err
	}
	r.fields.Reset()
	r.fieldsBuffer = r.fieldsBuffer[:0]
	for r.fields.next() {
		r.fieldsBuffer = append(r.fieldsBuffer, r.fields.field)
	}
	return r.fieldsBuffer, nil
}

func NewReader(r io.Reader) Reader {
	return Reader{
		fields: fields{
			buffer: bufferedReader{r: r, data: make([]byte, 0, 1024)},
		},
		fieldsBuffer: make([][]byte, 0, 16),
	}
}
