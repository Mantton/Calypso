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

func CreatePackage(dir string) (*Package, error) {
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

	p := NewPackage(dir, cfg)

	err = p.CollectModules()

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Package) CollectModules() error {

	// 2 - Read `src` folder of package
	SRC := path.Join(p.Path, "src")
	srcDir, err := os.Stat(SRC)

	if err != nil {
		return err
	}

	if !srcDir.IsDir() {
		return errors.New("\"src\" is not a directory")
	}
	// read entries in `src` directory
	entries, err := os.ReadDir(SRC)

	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	// Base module at the source directory
	base := NewModule(SRC, nil)
	p.AddModule(base)

	// iterate through src folder, collect files for base module, and subdirectories are top level modules of the package
	for _, entry := range entries {
		path := filepath.Join(SRC, entry.Name())

		// If directory in `src`, this is another standalone module
		if entry.IsDir() {
			mod, err := p.CollectModule(path, nil)

			if err != nil {
				return err
			}

			if mod == nil {
				continue
			}

			p.AddModule(mod)
		} else {
			base.AddFile(path)
		}
	}

	return nil
}

func (p *Package) CollectModule(path string, parent *Module) (*Module, error) {
	// read entries
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, nil
	}

	mod := NewModule(path, parent)
	for _, entry := range entries {
		path := filepath.Join(path, entry.Name())

		// Is SubModule/Directory
		if entry.IsDir() {
			submodule, err := p.CollectModule(path, mod)
			if err != nil {
				return nil, err
			}

			if mod == nil {
				continue
			}

			mod.AddModule(submodule)
		} else {
			mod.AddFile(path)
		}
	}

	return mod, nil
}
