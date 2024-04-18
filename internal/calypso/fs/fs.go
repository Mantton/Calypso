package fs

const CONFIG_FILE = "config.toml"

type FileSet struct {
	FilesPaths []string
}

type Module struct {
	Path         string
	Set          *FileSet
	SubModules   map[string]*Module
	ParentModule *Module
}

type Package struct {
	Modules  []*Module
	Config   *Config
	BasePath string
}

type LitePackage struct {
	Path   string
	Config *Config
}

type Config struct {
	Package struct {
		Name    string
		Version string
	}
	Dependencies map[string]*ConfigDependency
}

type ConfigDependency struct {
	Name    string
	Path    string
	Version string
	Alias   string
}

func (m *Module) AddSubmodule(s *Module) {
	if m.SubModules == nil {
		m.SubModules = make(map[string]*Module)
	}

	m.SubModules[s.Path] = s
	s.ParentModule = m
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
