package main

import (
	"testing"
)

type test struct {
	filePath string
	want     int
	wantErr  bool
}

var tests = []test{
	{"./tests/x.go", 6, false},
	{"./tests/x.lua", 2, false},
	{"./tests/x.js", 2, false},
	{"./tests/x.css", 3, false},
	{"./tests/x.html", 6, false},
	{"./tests/x", 0, true},
	{"./tests/notfound.go", 0, true},
}

func Test_sloc(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got, err := sloc(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("sloc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sloc() = %v, want %v", got, tt.want)
			}
		})
	}
}
