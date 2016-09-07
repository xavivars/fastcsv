package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var data, quotedData []byte

func init() {
	var err error
	data, err = ioutil.ReadFile("testdata/fl_insurance.csv")
	if err != nil {
		panic(err)
	}
	quotedData, err = ioutil.ReadFile("testdata/fl_insurance_quoted.csv")
	if err != nil {
		panic(err)
	}
}

func toStrings(bs [][]byte) []string {
	strs := make([]string, 0, len(bs))
	for _, b := range bs {
		strs = append(strs, string(b))
	}
	return strs
}

func quote(strs []string) []string {
	out := make([]string, 0, len(strs))
	for _, s := range strs {
		out = append(out, fmt.Sprintf("\"%s\"", s))
	}
	return out
}

func compareLine(line [][]byte, wanted ...string) error {
	if len(line) != len(wanted) {
		return fmt.Errorf(
			"Wanted [%s]; got [%s]",
			strings.Join(quote(wanted), ", "),
			strings.Join(quote(toStrings(line)), ", "),
		)
	}
	for i, s := range toStrings(line) {
		if s != wanted[i] {
			return fmt.Errorf(
				"Mismatch at item %d; wanted '%s'; got '%s'",
				i,
				wanted[i],
				s,
			)
		}
	}
	return nil
}

func testCSV(source string, wanted [][]string) error {
	r := NewReader(strings.NewReader(source))

	for i, wantedLine := range wanted {
		fields, err := r.Read()
		if err != nil {
			return fmt.Errorf("Unexpected error: %v", err)
		}
		if err := compareLine(fields, wantedLine...); err != nil {
			return fmt.Errorf("Mismatch on line %d: %v", i+1, err)
		}
	}

	if _, err := r.Read(); err != io.EOF {
		return fmt.Errorf("Wanted io.EOF; got: %v", err)
	}
	return nil
}

func TestReadOneRow(t *testing.T) {
	wanted := [][]string{{"abc", "def", "ghi"}}
	if err := testCSV("abc,def,ghi", wanted); err != nil {
		t.Fatal(err)
	}
}

func TestReadMultipleLines(t *testing.T) {
	wanted := [][]string{{"abc", "def"}, {"1234", "56"}}
	if err := testCSV("abc,def\n1234,56", wanted); err != nil {
		t.Fatal(err)
	}
}

func TestReadQuotedField(t *testing.T) {
	wanted := [][]string{{"abc", "123", "456"}}
	if err := testCSV("\"abc\",\"123\",\"456\"", wanted); err != nil {
		t.Fatal(err)
	}
}

func TestReadQuotedFieldMultipleLines(t *testing.T) {
	wanted := [][]string{{"abc", "123"}, {"def", "456"}}
	if err := testCSV("\"abc\",\"123\"\n\"def\",\"456\"", wanted); err != nil {
		t.Fatal(err)
	}
}

func TestReadQuotedFieldsWithComma(t *testing.T) {
	wanted := [][]string{{"a,b,c", "d,e,f"}}
	if err := testCSV("\"a,b,c\",\"d,e,f\"", wanted); err != nil {
		t.Fatal(err)
	}
}

func TestReadQuotedFieldsWithNewLine(t *testing.T) {
	if err := testCSV("\"a\nb\nc\"", [][]string{{"a\nb\nc"}}); err != nil {
		t.Fatal(err)
	}
}

func TestReadQuotedFieldsWithEscapedQuotes(t *testing.T) {
	if err := testCSV("\"a\"\"b\"", [][]string{{"a\"b"}}); err != nil {
		t.Fatal(err)
	}
}

func TestReadTrailingNewline(t *testing.T) {
	if err := testCSV("a,b,c\n", [][]string{{"a", "b", "c"}}); err != nil {
		t.Fatal(err)
	}
}

func TestReadEmptyMiddleLine(t *testing.T) {
	wanted := [][]string{{"a", "b"}, {""}, {"c", "d"}}
	if err := testCSV("a,b\n\nc,d", wanted); err != nil {
		t.Fatal(err)
	}
}

func TestReadCRLF(t *testing.T) {
	wanted := [][]string{{"a", "b", "c"}, {"d", "e", "f"}}
	if err := testCSV("a,b,c\r\nd,e,f", wanted); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkStdCsv(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewReader(data))
		for {
			if _, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkFastCsv(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewReader(bytes.NewReader(data))
		for {
			if _, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkStdCsvQuoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewReader(quotedData))
		for {
			if _, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkFastCsvQuoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewReader(bytes.NewReader(quotedData))
		for {
			if _, err := r.Read(); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatal(err)
			}
		}
	}
}
