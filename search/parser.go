package search

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

// parses csv data set and returns an array of strings
func parseDataset(path string) []string {

	f, err := os.Open(path)

	if err != nil {
		log.Fatal("Error opening dataset file", err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	var dataSet []string = make([]string, 0, 200)

	for {
		rec, err := csvReader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("Error reading csv file", err)
		}

		for _, data := range rec {
			dataSet = append(dataSet, data)
		}
	}

	return dataSet

}
