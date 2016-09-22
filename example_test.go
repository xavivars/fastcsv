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

		fmt.Printf("[%s]\n", string(bytes.Join(row, []byte(" | "))))
	}
	// Output:
	// [Language | Sponsor]
	// [golang | Google]
	// [swift | Apple]
	// [rust | Mozilla]
}

func ExampleReader_Next() {
	r := NewReader(strings.NewReader(`Language,Sponsor
golang,Google
swift,Apple
rust,Mozilla`))
	for r.Next() {
		fmt.Printf("[%s]\n", string(bytes.Join(r.Fields(), []byte(" | "))))
	}
	if err := r.Err(); err != nil {
		panic(err)
	}
	// Output:
	// [Language | Sponsor]
	// [golang | Google]
	// [swift | Apple]
	// [rust | Mozilla]
}
