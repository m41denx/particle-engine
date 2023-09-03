package hget

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func TaskPrint() error {
	downloading, err := os.ReadDir(filepath.Join(UserHome(), dataFolder))
	if err != nil {
		return err
	}

	folders := make([]string, 0)
	for _, d := range downloading {
		if d.IsDir() {
			folders = append(folders, d.Name())
		}
	}

	folderString := strings.Join(folders, "\n")
	Printf("Currently on going download: \n")
	fmt.Println(folderString)

	return nil
}

func Resume(task string) (*State, error) {
	return Read(task)
}
