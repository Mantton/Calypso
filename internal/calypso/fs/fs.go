package fs

import "fmt"

const CONFIG_FILE = "config.toml"

type FileSet struct {
	Paths []string
}

type Package struct {
	Modules []*Module
	Config  *Config
	Path    string
}

func NewPackage(path string, config *Config) *Package {
	return &Package{
		Path:   path,
		Config: config,
	}
}

func (p *Package) IsSTD() bool {
	return p.Path == GetSTDPath()
}

func (p *Package) AddModule(m *Module) {
	p.Modules = append(p.Modules, m)
}

func (p *Package) ID() string {
	return p.Config.ID()
}

// * Module
type Module struct {
	Path       string
	Files      *FileSet
	SubModules []*Module
	Parent     *Module
}

func NewModule(path string, parent *Module) *Module {
	return &Module{
		Path:   path,
		Parent: parent,
	}
}

func (m *Module) AddFile(f string) {
	if m.Files == nil {
		m.Files = &FileSet{
			Paths: []string{
				f,
			},
		}
	} else {
		m.Files.Paths = append(m.Files.Paths, f)
	}
}

func (m *Module) AddModule(sM *Module) {
	m.SubModules = append(m.SubModules, sM)
}

// Config
type Config struct {
	Package struct {
		Name    string
		Version string
	}
	Dependencies map[string]*ConfigDependency
}

func (c *Config) ID() string {
	return fmt.Sprintf("%s::%s", c.Package.Name, c.Package.Version)
}

// Config Dependency
type ConfigDependency struct {
	Name    string
	Path    string
	Version string
	Alias   string
}

func (c *ConfigDependency) ID() string {
	return fmt.Sprintf("%s::%s", c.Name, c.Version)
}

func (c *Config) FindDependency(n string) *ConfigDependency {

	if dep, ok := c.Dependencies[n]; ok {
		return dep
	}

	for _, dep := range c.Dependencies {
		if dep.Alias == n {
			return dep
		}
	}

	return nil
}

func (c *ConfigDependency) FilePath() string {
	if len(c.Path) != 0 {
		return c.Path
	}

	panic("TODO: downloaded package")
}
