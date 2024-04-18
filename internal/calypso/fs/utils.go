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

func CreateLitePackage(dir string, isProject bool) (*LitePackage, error) {
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

	return &LitePackage{
		Path:   dir,
		Config: cfg,
	}, nil

}

func CollectModule(path string, subs bool) (*Module, error) {

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
		Path:       path,
	}

	for _, entry := range entries {
		p := filepath.Join(path, entry.Name())

		// Is SubModule/Directory
		if entry.IsDir() {
			if subs {
				continue
			}
			sub, err := CollectModule(p, true)

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
