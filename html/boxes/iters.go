package boxes

type boxIterator interface {
	Next() bool
	Box() Box
}

// implements boxIterator
type boxSlice struct {
	data []Box
	pos  int
}

func newBoxIter(boxes []Box) *boxSlice { return &boxSlice{data: boxes} }

func (s boxSlice) Next() bool { return s.pos < len(s.data) }

func (s *boxSlice) Box() Box {
	b := s.data[s.pos]
	s.pos++
	return b
}

func collectBoxes(iter boxIterator) []Box {
	var out []Box
	for iter.Next() {
		out = append(out, iter.Box())
	}
	return out
}
