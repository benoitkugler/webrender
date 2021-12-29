package svg

import (
	"reflect"
	"testing"
)

func moveToF(x, y Fl) pathItem {
	return pathItem{op: moveTo, args: [3]point{{x, y}}}
}

func lineToF(x, y Fl) pathItem {
	return pathItem{op: lineTo, args: [3]point{{x, y}}}
}

func cubicToF(x1, y1, x2, y2, x3, y3 Fl) pathItem {
	return pathItem{op: cubicTo, args: [3]point{
		{x1, y1}, {x2, y2}, {x3, y3},
	}}
}

func Test_parsePath(t *testing.T) {
	tests := []struct {
		args    string
		want    []pathItem
		wantErr bool
	}{
		{
			args: "M200,300 L400,50 L600,300 L800,550 L1000,300",
			want: []pathItem{
				moveToF(200, 300),
				lineToF(400, 50),
				lineToF(600, 300),
				lineToF(800, 550),
				lineToF(1000, 300),
			},
		},
		{
			args: "M100,200 C100,100 250,100 250,200",
			want: []pathItem{
				moveToF(100, 200),
				cubicToF(100, 100, 250, 100, 250, 200),
			},
		},
		{
			args:    "M100, ",
			wantErr: true,
		},
	}
	var c pathParser
	for _, tt := range tests {
		got, err := c.parsePath(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("parsePath() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parsePath() = %v, want %v", got, tt.want)
		}
	}
}

func TestParsePathCrash(t *testing.T) {
	paths := []string{
		"M300,200 h-150 a150,150 0 1,0 150,-150 z",
		"M275,175 v-150 a150,150 0 0,0 -150,150 z",
		`M600,350 l 50,-25 
			a25,25 -30 0,1 50,-25 l 50,-25 
			a25,50 -30 0,1 50,-25 l 50,-25 
			a25,75 -30 0,1 50,-25 l 50,-25 
			a25,100 -30 0,1 50,-25 l 50,-25`,
		"M200,300 Q400,50 600,300 T1000,300",
		"M100,200 C100,100 250,100 250,200 S400,300 400,200",
	}
	var c pathParser
	for _, path := range paths {
		_, err := c.parsePath(path)
		if err != nil {
			t.Fatal(err)
		}
	}
}
