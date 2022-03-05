package xpath

import (
	"log"
	"os"
	"path/filepath"
)

func GetExecutablePath() string {
	path, err := os.Executable()
	if err != nil {
		log.Print(err)
		path = os.Args[0]
	}
	dir := filepath.Dir(path)
	return dir
}
