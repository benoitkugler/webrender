package tree

import (
	"fmt"
	"testing"
)

func TestResumeStackEqual(t *testing.T) {
	for _, data := range []struct {
		s1, s2   ResumeStack
		expected bool
	}{
		{ResumeStack{}, nil, true},
		{ResumeStack{2: nil}, ResumeStack{2: ResumeStack{}}, true},
		{ResumeStack{3: nil}, ResumeStack{2: ResumeStack{}}, false},
		{ResumeStack{2: nil}, nil, false},
		{ResumeStack{2: nil, 3: nil}, ResumeStack{3: nil, 2: nil}, true},
		{ResumeStack{2: nil, 3: nil}, ResumeStack{3: nil}, false},
	} {
		if data.s1.Equals(data.s2) != data.expected {
			t.Fatalf("unexpected comparison for %v and %v", data.s1, data.s2)
		}
		fmt.Println(data.s1, data.s2)
	}
}
