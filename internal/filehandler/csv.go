package filehandler

import (
	"encoding/csv"
	"os"


	"fmt"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) ReadSchema(filename string, delimiter rune) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = delimiter

	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	return headers, nil
}

func (s *FileService) ReadData(filename string, delimiter rune, selectedColumns []string) (<-chan []map[string]string, <-chan error) {
	dataChan := make(chan []map[string]string)
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		file, err := os.Open(filename)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = delimiter

		headers, err := reader.Read()
		if err != nil {
			errChan <- err
			return
		}

		columnIndices := make([]int, 0, len(selectedColumns))
		for _, col := range selectedColumns {
			for i, h := range headers {
				if h == col {
					columnIndices = append(columnIndices, i)
					break
				}
			}
		}

		var batch []map[string]string
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			entry := make(map[string]string)
			for i, colIdx := range columnIndices {
				if colIdx < len(record) {
					entry[selectedColumns[i]] = record[colIdx]
				}
			}

			batch = append(batch, entry)
			if len(batch) >= 1000 { // Batch size
				dataChan <- batch
				batch = nil
			}
		}

		if len(batch) > 0 {
			dataChan <- batch
		}
	}()

	return dataChan, errChan
}

func (s *FileService) WriteData(filename string, delimiter rune, headers []string, data <-chan []map[string]interface{}) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		file, err := os.Create(filename)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Comma = delimiter

		if err := writer.Write(headers); err != nil {
			errChan <- err
			return
		}

		for batch := range data {
			for _, record := range batch {
				row := make([]string, len(headers))
				for i, h := range headers {
					if val, ok := record[h].(string); ok {
						row[i] = val
					} else {
						row[i] = fmt.Sprintf("%v", record[h])
					}
				}
				if err := writer.Write(row); err != nil {
					errChan <- err
					return
				}
			}
			writer.Flush()
		}
	}()

	return errChan
}