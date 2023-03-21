package output

import (
	"encoding/csv"
	"fmt"
	"os"
)

func CreateDataCsvFile(csvFileName string, totalRecords int, funcToGenerateDataRow func() ([]string, error)) error {
	f, err := os.Create(csvFileName)

	if err != nil {
		return fmt.Errorf("failed to create file %s %v", csvFileName, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	for i := 0; i < totalRecords; i++ {
		row, err := funcToGenerateDataRow()
		if err != nil {
			return fmt.Errorf("failed to generate data row: %v", err)
		}

		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write the data row to csv file: %v", err)
		}
	}
	return nil
}
