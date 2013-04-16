package staticcontent

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
)

var static_content_config = flag.String("static_content_config", "",
	"Path to the static content config csv file, for auto-configuring custom"+
		"static paths. See staticcontent/example_paths.csv for examples.")

// Adds paths to the static content server based on the shared configuration
// file. Any paths that cannot be interpreted or found will be ignored.
func AddCsvPaths(s *Server) error {
	if len(*static_content_config) == 0 {
		log.Println("No static content config found; not mapping any " +
			"directories. Set --static_content_config to have static content " +
			"hosted.")
		return nil
	}
	file, err := os.Open(*static_content_config)
	if err != nil {
		return err
	}
	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(record) != 2 {
			log.Println("Malformed mapping in static content csv file:", record)
			continue
		}
		s.Install(record[0], record[1])
	}
	return nil
}
