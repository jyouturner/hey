package input

import (
	"encoding/csv"
	"io"
)

// CsvInputReader handles reading from CSV file, implementing the InputReader interface
type CsvInputReader struct {
	ignoreFirstLine bool
	reader          *csv.Reader
}

func NewCsvInputReader(f io.Reader, igoreFirstLine bool) *CsvInputReader {
	csvReader := csv.NewReader(f)

	s := CsvInputReader{
		reader:          csvReader,
		ignoreFirstLine: igoreFirstLine,
	}

	return &s
}

func (s *CsvInputReader) Read() ([]string, error) {
	rec, err := s.reader.Read()
	if err != nil {
		if err == io.EOF {
			//end of the file
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return rec, nil
	}

}
