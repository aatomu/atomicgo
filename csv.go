package atomicgo

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"strings"
)

type CsvOptions struct {
	Comma            rune // 区切り文字
	Comment          rune // コメントの先頭
	FieldsPerRecord  int  // 各行のフィールド数。多くても少なくてもエラーになる
	LazyQuotes       bool // true の場合、"" が値の途中に "180"cm のようになっていてもエラーにならない
	TrimLeadingSpace bool // true の場合は、先頭の空白文字を無視する
}

func CsvReadStr(data string, option CsvOptions) (c [][]string, err error) {
	r := parseReader(strings.NewReader(data), option)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return [][]string{}, err
		}
		c = append(c, record)
	}
	return
}

func CsvReadStrFlash(data string, option CsvOptions) (c [][]string, err error) {
	r := parseReader(strings.NewReader(data), option)
	return r.ReadAll()
}

func CsvReadFile(data string, option CsvOptions) (c [][]string, err error) {
	f, err := os.Open("file.csv")
	if err != nil {
		return
	}

	r := parseReader(f, option)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return [][]string{}, err
		}
		c = append(c, record)
	}
	return
}

func CsvReadFileFlash(data string, option CsvOptions) (c [][]string, err error) {
	f, err := os.Open("file.csv")
	if err != nil {
		return
	}

	r := parseReader(f, option)
	return r.ReadAll()
}

func parseReader(i io.Reader, ops CsvOptions) (c *csv.Reader) {
	r := csv.NewReader(i)

	if ops.Comma != 0 {
		r.Comma = ops.Comma
	}
	if ops.Comment != 0 {
		r.Comment = ops.Comment
	}
	if ops.FieldsPerRecord != 0 {
		r.FieldsPerRecord = ops.FieldsPerRecord
	}
	r.LazyQuotes = ops.LazyQuotes
	r.TrimLeadingSpace = ops.TrimLeadingSpace
	return r
}

func CsvFormat(data [][]string, comma rune) (c string, err error) {
	buf := new(bytes.Buffer)
	t := csv.NewWriter(buf)
	t.Comma = comma
	for _, d := range data {
		err = t.Write(d)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func CsvFormatFlash(data [][]string, comma rune) (c string, err error) {
	buf := new(bytes.Buffer)
	t := csv.NewWriter(buf)
	t.Comma = comma
	err = t.WriteAll(data)
	return buf.String(), err
}
