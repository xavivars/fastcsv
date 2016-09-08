package csv

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func ExampleReader_Read() {
	r := NewReader(strings.NewReader(`Language,Sponsor
golang,Google
swift,Apple
rust,Mozilla`))
	for {
		row, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		fmt.Println(string(bytes.Join(row, []byte(", "))))
	}
}

func ExampleReader_iterating() {
	r := NewReader(strings.NewReader(`Language,Sponsor
golang,Google
swift,Apple
rust,Mozilla`))
	for r.Next() {
		fmt.Println(string(bytes.Join(r.Fields(), []byte(", "))))
	}
	if err := r.Err(); err != nil {
		panic(err)
	}
}
