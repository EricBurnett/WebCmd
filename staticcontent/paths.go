package staticcontent

import (
	"bytes"
	"encoding/csv"
	"github.com/EricBurnett/WebCmd/resources"
	"io"
	"log"
)

var STATIC_CONTENT_FILE = "staticcontent/paths.csv"

// Adds paths to the static content server based on the shared configuration
// file. Any paths that cannot be interpreted or found will be ignored.
func AddCsvPaths(s *Server) error {
	data, err := resources.Load(STATIC_CONTENT_FILE)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bytes.NewReader(data))
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(record) != 2 {
			log.Println("Malformed record in static content csv file:", record)
			continue
		}
		s.Install(record[0], record[1])
	}
	return nil
}
