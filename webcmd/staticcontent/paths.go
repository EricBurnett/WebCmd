package staticcontent

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

var STATIC_CONTENT_FILE = "staticcontent/paths.csv"

// Adds paths to the static content server based on the shared configuration
// file. Any paths that cannot be interpreted or found will be ignored.
func AddCsvPaths(s *Server) error {
	file, err := os.Open(STATIC_CONTENT_FILE)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
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
