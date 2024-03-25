package utils

type FileSet struct {
	FilesPaths []string
}

type Module struct {
	Set        *FileSet
	SubModules map[string]*Module
}

type Project struct {
	Modules map[string]*Module
}
