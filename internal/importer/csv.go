package importer

import (
	"encoding/csv"
	"io"
)

type Prospect struct {
	Company          string
	Email            string
	IdentifiedProblem string
}

func Read(r io.Reader) ([]Prospect, error) {
	reader := csv.NewReader(r)

	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var prospects []Prospect

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		prospects = append(prospects, Prospect{
			Company:          record[0],
			Email:            record[1],
			IdentifiedProblem: record[2],
		})
	}

	return prospects, nil
}
