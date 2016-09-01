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

func toStrings(bs [][]byte) []string {
	strs := make([]string, 0, len(bs))
	for _, b := range bs {
		strs = append(strs, string(b))
	}
	return strs
}

func compareLine(line [][]byte, wanted ...string) error {
	if len(line) != len(wanted) {
		return fmt.Errorf(
			"Wanted [%s]; got [%s]",
			strings.Join(wanted, ", "),
			strings.Join(toStrings(line), ", "),
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

func TestReadOneRow(t *testing.T) {
	r := NewReader(strings.NewReader("abc,def,ghi"))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := compareLine(fields, "abc", "def", "ghi"); err != nil {
		t.Fatal(err)
	}

	if _, err = r.Read(); err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func TestReadMultipleLines(t *testing.T) {
	r := NewReader(strings.NewReader("abc,def\n1234,56"))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := compareLine(fields, "abc", "def"); err != nil {
		t.Fatal(err)
	}

	fields, err = r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if err := compareLine(fields, "1234", "56"); err != nil {
		t.Fatal(err)
	}

	if _, err := r.Read(); err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func TestReadQuotedField(t *testing.T) {
	r := NewReader(strings.NewReader("\"abc\",\"123\",\"456\""))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if err := compareLine(fields, "abc", "123", "456"); err != nil {
		t.Fatal(err)
	}

	_, err = r.Read()
	if err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func TestReadQuotedFieldMultipleLines(t *testing.T) {
	r := NewReader(strings.NewReader("\"abc\",\"123\"\n\"def\",\"456\""))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if err := compareLine(fields, "abc", "123"); err != nil {
		t.Fatal(err)
	}

	fields, err = r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if err := compareLine(fields, "def", "456"); err != nil {
		t.Fatal(err)
	}

	_, err = r.Read()
	if err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func TestReadQuotedFieldsWithComma(t *testing.T) {
	r := NewReader(strings.NewReader("\"a,b,c\",\"d,e,f\""))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if err := compareLine(fields, "a,b,c", "d,e,f"); err != nil {
		t.Fatal(err)
	}

	_, err = r.Read()
	if err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func TestReadQuotedFieldsWithNewLine(t *testing.T) {
	r := NewReader(strings.NewReader("\"a\nb\nc\""))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := compareLine(fields, "a\nb\nc"); err != nil {
		t.Fatal(err)
	}

	_, err = r.Read()
	if err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func TestReadQuotedFieldsWithEscapedQuotes(t *testing.T) {
	r := NewReader(strings.NewReader("\"a\"\"b\""))

	fields, err := r.Read()
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if err := compareLine(fields, "a\"b"); err != nil {
		t.Fatal(err)
	}

	_, err = r.Read()
	if err != io.EOF {
		t.Fatal("Wanted io.EOF; got:", err)
	}
}

func BenchmarkStdCsv(b *testing.B) {
	data, err := ioutil.ReadFile("fl_insurance.csv")
	if err != nil {
		b.Fatal(err)
	}
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

func BenchmarkMyCsv(b *testing.B) {
	data, err := ioutil.ReadFile("fl_insurance.csv")
	if err != nil {
		b.Fatal(err)
	}
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
	data, err := ioutil.ReadFile("fl_insurance_quoted.csv")
	if err != nil {
		b.Fatal(err)
	}
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

func BenchmarkMyCsvQuoted(b *testing.B) {
	data, err := ioutil.ReadFile("fl_insurance_quoted.csv")
	if err != nil {
		b.Fatal(err)
	}
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
