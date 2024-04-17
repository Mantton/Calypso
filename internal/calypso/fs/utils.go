package fs

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func ParseConfig(path string) (*Config, error) {

	var cfg *Config

	content, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(content, &cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func CollectPackage(path string, isProject bool) (*Package, error) {

	// Building a project
	if isProject {
		return buildPackage(path)
	}

	// Building a file or two
	src, err := buildModule(path)

	if err != nil {
		return nil, err
	}

	return &Package{
		Source: src,
	}, nil
}

func buildPackage(dir string) (*Package, error) {
	// 1 - Read Config File

	cfgPath := path.Join(dir, CONFIG_FILE)
	_, err := os.Stat(cfgPath)

	if err != nil {
		return nil, err
	}

	cfg, err := ParseConfig(cfgPath)

	if err != nil {
		return nil, err
	}

	// 2 - Read src folder
	srcPath := path.Join(dir, "src")
	srcDir, err := os.Stat(srcPath)

	if err != nil {
		return nil, err
	}

	if !srcDir.IsDir() {
		return nil, errors.New("\"src\" is not a directory")
	}

	// Collect Modules & SubModules

	module, err := buildModule(srcPath)

	if err != nil {
		return nil, err
	}

	pkg := &Package{
		Source: module,
		Config: cfg,
	}

	return pkg, nil
}

func buildModule(path string) (*Module, error) {

	// read entries
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	// declare
	set := &FileSet{}
	var submodules map[string]*Module

	mod := &Module{
		Set:        set,
		SubModules: submodules,
		FolderName: path,
	}

	for _, entry := range entries {
		p := filepath.Join(path, entry.Name())

		// Is SubModule/Directory
		if entry.IsDir() {
			sub, err := buildModule(p)

			if err != nil {
				return nil, err
			}

			mod.AddSubmodule(sub)
		} else {
			// is file, add to fileset
			set.FilesPaths = append(set.FilesPaths, p)
		}
	}

	return mod, nil
}
