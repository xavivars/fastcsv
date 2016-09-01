package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("fl_insurance.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	outFile, err := os.Create("fl_insurance_quoted.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	for {
		row, err := r.Read()
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}
		for i := range row {
			row[i] = fmt.Sprintf("\"%s\"", row[i])
		}
		fmt.Fprintln(outFile, strings.Join(row, ","))
	}
}
