package commands

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mantton/calypso/internal/calypso/builder"
	"github.com/mantton/calypso/internal/calypso/commands/utils"
)

const CONFIG_FILE = "config.toml"

func build(paths []string) error {

	switch len(paths) {
	case 0:
		dir, err := os.Getwd()

		if err != nil {
			return err
		}

		return buildFromDirectory(dir)
	case 1:
		return buildFromPath(paths[0])
	default:
		return buildFromFileList(paths)
	}
}

func buildFromPath(path string) error {
	// Is File or Directory

	f, err := os.Stat(path)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if f.IsDir() {
		return buildFromDirectory(path)
	}

	return buildFromFile(path)
}

func buildFromFile(pth string) error {
	fmt.Println("Build from file", pth)
	return nil
}

/*
Builds a list of provided files

RULES:
- All Files must be in the same directory
- All Files must belong to the same module
*/
func buildFromFileList(paths []string) error {
	set := &utils.FileSet{}

	// Satisfy Rule 1

	// ensure all paths are files & collect directories of files
	dirs := make(map[string]struct{})
	for _, path := range paths {
		file, err := os.Stat(path)

		if err != nil {
			return err
		}

		if file.IsDir() {
			return fmt.Errorf("\"%s\" is a directory", file.Name())
		}

		dirs[filepath.Dir(path)] = struct{}{}

		set.FilesPaths = append(set.FilesPaths, path)
	}

	// map acts as a set in this case where we check that the dir lenght is just one, meaning one directory
	if len(dirs) > 1 {
		names := []string{}
		for dir := range dirs {
			names = append(names, dir)
		}
		s := strings.Join(names, ", ")
		return fmt.Errorf("all files must be in the same directory, got: %s", s)
	}

	return builder.CompileFileSet(set)
}

func buildFromDirectory(dir string) error {
	// 1 - Read Config File
	_, err := os.ReadFile(path.Join(dir, CONFIG_FILE))

	if err != nil {
		return err
	}

	// 2 - Read src folder
	srcDir, err := os.Stat(path.Join(dir, "src"))

	if err != nil {
		return err
	}

	if !srcDir.IsDir() {
		return errors.New("\"src\" is not a directory")
	}

	return nil
}
