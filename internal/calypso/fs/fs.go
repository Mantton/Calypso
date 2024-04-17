package fs

const CONFIG_FILE = "config.toml"

type FileSet struct {
	FilesPaths []string
}

type Module struct {
	FolderName   string
	Set          *FileSet
	SubModules   map[string]*Module
	ParentModule *Module
}

type Package struct {
	Source *Module
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
}

func (m *Module) AddSubmodule(s *Module) {
	if m.SubModules == nil {
		m.SubModules = make(map[string]*Module)
	}

	m.SubModules[s.FolderName] = s
	s.ParentModule = m
}
