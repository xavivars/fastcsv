package csv

import "io"

type bufferedReader struct {
	r      io.Reader
	data   []byte
	cursor int
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
	// skip past the initial quote rune
	buffer.cursor++
	start := buffer.cursor

	writeCursor := buffer.cursor
	last := byte(0)
	for {
		// next byte
		if buffer.cursor >= len(buffer.data) {
			if err := buffer.more(); err != nil {
				return buffer.data[start:writeCursor], true, err
			}
		}
		ch := buffer.data[buffer.cursor]
		buffer.cursor++

		// handle byte
		switch ch {
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
				last = ch
				continue
			}
		}
		writeCursor++
		last = ch
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

	if first := fs.buffer.data[fs.buffer.cursor]; first == '"' {
		fs.field, fs.hitEOL, fs.err = nextQuotedField(&fs.buffer)
		return fs.err == nil || fs.err == io.EOF
	}
	return fs.nextUnquotedField()
}

type Reader struct {
	fields       fields
	fieldsBuffer [][]byte
}

// Scans in the next row
func (r *Reader) Next() bool {
	if r.fields.err != nil {
		return false
	}
	r.fields.reset()
	r.fieldsBuffer = r.fieldsBuffer[:0]
	for r.fields.next() {
		r.fieldsBuffer = append(r.fieldsBuffer, r.fields.field)
	}

	// CRLF support: if there are fields in this row, and the last field ends
	// with `\r`, then it must have been part of a CRLF line ending, so drop
	// the `\r`.
	if len(r.fieldsBuffer) > 0 {
		lastField := r.fieldsBuffer[len(r.fieldsBuffer)-1]
		if len(lastField) > 0 && lastField[len(lastField)-1] == '\r' {
			lastField = lastField[:len(lastField)-1]
			r.fieldsBuffer[len(r.fieldsBuffer)-1] = lastField
		}
	}

	// Handle CSVs that end with a blank last line
	if len(r.fieldsBuffer) == 0 {
		r.fields.err = io.EOF
		return false
	}

	return true
}

// Returns the last row of fields encountered. These fields are only valid
// until the next call to Next() or Read().
func (r *Reader) Fields() [][]byte {
	return r.fieldsBuffer
}

// Return the last error encountered; returns nil if no error was encountered
// or if the last error was io.EOF.
func (r *Reader) Err() error {
	if r.fields.err != io.EOF {
		return r.fields.err
	}
	return nil
}

// Read and return the next row and/or any errors encountered. The byte slices
// are only valid until the next call to Next() or Read(). Returns nil, io.EOF
// when the file is consumed.
func (r *Reader) Read() ([][]byte, error) {
	if r.Next() {
		return r.fieldsBuffer, nil
	}
	return nil, r.fields.err
}

// Constructs a new Reader from a source CSV io.Reader
func NewReader(r io.Reader) Reader {
	return Reader{
		fields: fields{
			buffer: bufferedReader{r: r, data: make([]byte, 0, 1024)},
		},
		fieldsBuffer: make([][]byte, 0, 16),
	}
}
