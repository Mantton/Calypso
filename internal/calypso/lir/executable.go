package lir

func NewExecutable() *Executable {
	return &Executable{
		Modules: make(map[string]*Module),
	}
}
