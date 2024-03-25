package utils

type File struct {
	Path string
}

type FileSet struct {
	Files []File
}

type Module struct {
	Set        *FileSet
	SubModules map[string]*Module
}

type Project struct {
	Modules map[string]*Module
}
