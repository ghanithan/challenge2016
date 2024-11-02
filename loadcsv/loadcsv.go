package loadcsv

import (
	"encoding/csv"
	"fmt"
	"os"
)

func LoadCsv(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error in opening file: %s", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	csv, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error in parsing the CSV file: %s", err)
	}
	//fmt.Println(csv)
	return csv, nil
}
