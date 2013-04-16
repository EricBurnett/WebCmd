package resources

import (
	"bitbucket.org/kardianos/osext"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
)

var resource_path = flag.String("resource_path", "",
	"Path to the program resources. If not specified, assumed to be in the "+
		"standard location relative to an installed binary, i.e. "+
		"<binary path>/../src/github.com/EricBurnett/WebCmd")

func Load(path string) ([]byte, error) {
	resourcePath, err := ResourcePath()
	if err != nil {
		log.Println("Unable to determine resource load path!")
		return nil, err
	}
	file, err := ioutil.ReadFile(filepath.Join(resourcePath, path))
	if err == nil {
		log.Println("Successfully loaded resource", path, "from", resourcePath)
	} else {
		log.Println("Unable to read", path, "in", resourcePath, " - resource not loaded.")
	}
	return file, err
}

func ResourcePath() (string, error) {
	if len(*resource_path) > 0 {
		return *resource_path, nil
	}

	executablePath, err := osext.ExecutableFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(executablePath, "..", "src", "github.com", "EricBurnett", "WebCmd"), nil
}
