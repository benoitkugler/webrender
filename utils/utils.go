package utils

var Has = struct{}{}

type Set map[string]struct{}

func (s Set) Add(key string) {
	s[key] = Has
}

func (s Set) Extend(keys []string) {
	for _, key := range keys {
		s[key] = Has
	}
}

func (s Set) Has(key string) bool {
	_, in := s[key]
	return in
}

// Copy returns a deepcopy.
func (s Set) Copy() Set {
	out := make(Set, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func (s Set) IsNone() bool { return s == nil }

func (s Set) Equal(other Set) bool {
	if len(s) != len(other) {
		return false
	}
	for i := range s {
		if _, in := other[i]; !in {
			return false
		}
	}
	return true
}

func NewSet(values ...string) Set {
	s := make(Set, len(values))
	for _, v := range values {
		s.Add(v)
	}
	return s
}
