package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mantton/calypso/internal/calypso/compile"
	"github.com/mantton/calypso/internal/calypso/fs"
)

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
	// set := &fs.FileSet{FilesPaths: []string{pth}}
	panic("unimplemented")
}

/*
Builds a list of provided files

RULES:
- All Files must be in the same directory
- All Files must belong to the same module
*/
func buildFromFileList(paths []string) error {
	set := &fs.FileSet{}

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

		set.Paths = append(set.Paths, path)
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

	panic("unimplemented")
	// return builder.CompileFileSet(set, typechecker.USER)
}

func buildFromDirectory(path string) error {
	return compile.CompilePackage(path)
}
