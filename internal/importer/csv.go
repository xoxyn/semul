package importer

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/xoxyn/semul/internal/domain"
)

func Read(r io.Reader) ([]domain.Prospect, error) {
	reader := csv.NewReader(r)

	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var prospects []domain.Prospect

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) != 3 {
			return nil, fmt.Errorf("expected 3 columns, got %d", len(record))
		}

		prospects = append(prospects, domain.Prospect{
			Company:           record[0],
			Email:             record[1],
			IdentifiedProblem: record[2],
		})
	}

	return prospects, nil
}
