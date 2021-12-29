package svg

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/webrender/css/parser"
)

func Test_newPainter(t *testing.T) {
	tests := []struct {
		args    string
		want    painter
		wantErr bool
	}{
		{
			"red",
			painter{"", parser.ColorKeywords["red"].RGBA, true},
			false,
		},
		{
			"",
			painter{"", parser.RGBA{}, false},
			false,
		},
		{
			"none",
			painter{"", parser.RGBA{}, false},
			false,
		},
		{
			"black",
			painter{"", parser.RGBA{A: 1}, true},
			false,
		},
		{
			"url(ddd",
			painter{},
			true,
		},
		{
			"url(#myPaint)",
			painter{"myPaint", parser.RGBA{}, true},
			false,
		},
		{
			"url(#myPaint)  green",
			painter{"myPaint", parser.ColorKeywords["green"].RGBA, true},
			false,
		},
	}
	for _, tt := range tests {
		got, err := newPainter(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("newPainter() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("newPainter() = %v, want %v", got, tt.want)
		}
	}
}

func Test_clampModulo(t *testing.T) {
	type args struct {
		offset Fl
		total  Fl
	}
	tests := []struct {
		args args
		want Fl
	}{
		{args{10, 20}, 10},
		{args{11.2, 22.3}, 11.2},
		{args{-1, 22.3}, 21.3},
		{args{-10.5, 22.3}, 11.799999},
		{args{-30, 22.3}, 14.599998},
	}
	for _, tt := range tests {
		if got := clampModulo(tt.args.offset, tt.args.total); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("clampModulo() = %v, want %v", got, tt.want)
		}
	}
}
