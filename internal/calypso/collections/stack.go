package collections

type Stack[T any] struct {
	elements []T
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}

func (s *Stack[T]) Length() int {
	return len(s.elements)
}

func (s *Stack[T]) Push(v T) {
	s.elements = append(s.elements, v)
}

func (s *Stack[T]) Get(idx int) (T, bool) {
	if s.IsEmpty() || idx > s.Length()-1 {
		var none T
		return none, false
	}
	return s.elements[idx], true
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var none T
		return none, false
	} else {
		index := len(s.elements) - 1    // Index of Top Element
		element := (s.elements)[index]  // Get Element to be returned
		s.elements = s.elements[:index] // array is now simply the array from the start up till the prev index
		return element, true
	}
}

func (s *Stack[T]) Head() (T, bool) {

	if s.IsEmpty() {
		var none T
		return none, false
	}
	index := len(s.elements) - 1   // Index of Top Element
	element := (s.elements)[index] // Get Element to be returned
	return element, true
}
