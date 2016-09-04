package csv

import "io"

type bufferedReader struct {
	r      io.Reader
	data   []byte
	cursor int
}

func (b *bufferedReader) peek() (byte, error) {
	if b.cursor >= len(b.data) {
		if err := b.more(); err != nil {
			return byte(0), err
		}
	}
	return b.data[b.cursor], nil
}

func (b *bufferedReader) next() (byte, error) {
	if b.cursor >= len(b.data) {
		if err := b.more(); err != nil {
			return byte(0), err
		}
	}
	ch := b.data[b.cursor]
	b.cursor++
	return ch, nil
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

func (fs *fields) reset() {
	fs.buffer.reset()
	fs.field = nil
	fs.fieldStart = 0
	fs.hitEOL = false
}

func (fs *fields) nextUnquotedField() bool {
	const sizeEOL = 1
	const sizeDelim = 1
	for {
		// next byte
		if fs.buffer.cursor >= len(fs.buffer.data) {
			if err := fs.buffer.more(); err != nil {
				if err == io.EOF {
					start := fs.fieldStart
					end := fs.buffer.cursor
					fs.field = fs.buffer.data[start:end]
					fs.hitEOL = true
					fs.err = err
					return true
				}
				fs.err = err
				return false
			}
		}
		ch := fs.buffer.data[fs.buffer.cursor]
		fs.buffer.cursor++

		// handle byte
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
	buffer.next()
	start := buffer.cursor

	writeCursor := buffer.cursor
	last := byte(0)
	for {
		r, err := buffer.next()
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
		writeCursor++
		last = r
		// copy the current rune onto writeCursor if writeCursor !=
		// buffer.cursor
		if writeCursor != buffer.cursor {
			copy(
				buffer.data[writeCursor:writeCursor+1],
				buffer.data[buffer.cursor:buffer.cursor+1],
			)
		}
	}
}

func (fs *fields) next() bool {
	if fs.hitEOL {
		return false
	}
	if fs.buffer.cursor >= len(fs.buffer.data) {
		if err := fs.buffer.more(); err != nil {
			fs.err = err
			return false
		}
	}

	if fs.buffer.data[fs.buffer.cursor] == '"' {
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
	r.fields.reset()
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
