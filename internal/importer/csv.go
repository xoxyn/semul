package importer

import (
	"encoding/csv"
	"io"
	"log"
	"strings"
)

func Read() []string {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
	r := csv.NewReader(strings.NewReader(in))
	var result []string

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		result = append(result, record...)
	}
	return result
}
