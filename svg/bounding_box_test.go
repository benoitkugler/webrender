package svg

import (
	"reflect"
	"testing"
)

func Test_path_boundingBox(t *testing.T) {
	tests := []struct {
		p     path
		want  Rectangle
		want1 bool
	}{
		{
			// empty path
			want1: false,
		},
		{
			p: path{
				moveToF(10, 20),
				lineToF(30, 50),
			},
			want:  Rectangle{10, 20, 20, 30},
			want1: true,
		},
	}
	for _, tt := range tests {
		got, got1 := tt.p.boundingBox(nil, drawingDims{})
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("path.boundingBox() got = %v, want %v", got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("path.boundingBox() got1 = %v, want %v", got1, tt.want1)
		}
	}
}
