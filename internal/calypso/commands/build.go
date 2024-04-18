package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mantton/calypso/internal/calypso/builder"
	"github.com/mantton/calypso/internal/calypso/compile"
	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/typechecker"
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
	set := &fs.FileSet{FilesPaths: []string{pth}}
	return builder.CompileFileSet(set, typechecker.USER)
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

	return builder.CompileFileSet(set, typechecker.USER)
}

func buildFromDirectory(path string) error {
	// collect file paths & group into modules and submodules
	pkg, err := fs.CreateLitePackage(path, true)

	if err != nil {
		return err
	}

	// compile
	return compile.CompilePackage(pkg)
}
