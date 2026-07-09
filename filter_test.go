package main

import (
	"os"
	"slices"
	"testing"
)

type fakeInfo struct {
	os.FileInfo
	name string
}

func (f fakeInfo) Name() string { return f.name }

func mkEntries(names []string) (es []dirEntry) {
	for _, n := range names {
		es = append(es, dirEntry{fi: fakeInfo{name: n}})
	}
	return es
}

func entryNames(es []dirEntry) (ns []string) {
	for _, e := range es {
		ns = append(ns, e.fi.Name())
	}
	return ns
}

func TestFilterEntries(t *testing.T) {
	dNames := []string{"src", "Foobox", "docs"}
	fNames := []string{"README.md", "main.go", "main_test.go", "foobar.txt", "Foo.jpg", "notes"}
	dirs, files := mkEntries(dNames), mkEntries(fNames)
	cases := []struct {
		pat          string
		wantD, wantF []string
	}{
		{"", dNames, fNames},
		{"*.go", nil, []string{"main.go", "main_test.go"}},
		{"*.*", nil, []string{"README.md", "main.go", "main_test.go", "foobar.txt", "Foo.jpg"}},
		{"main*", nil, []string{"main.go", "main_test.go"}},
		{"*.JPG", nil, []string{"Foo.jpg"}},                            // glob, case-insensitive
		{"notes", nil, []string{"notes"}},                              // exact match anywhere, no fallback
		{"src", []string{"src"}, nil},                                  // exact dir match
		{"foo", []string{"Foobox"}, []string{"foobar.txt", "Foo.jpg"}}, // keyword fallback
		{"readme", nil, []string{"README.md"}},
		{`\.GO$`, nil, []string{"main.go", "main_test.go"}}, // regex, case-insensitive
		{"^(foo|main)", []string{"Foobox"}, []string{"main.go", "main_test.go", "foobar.txt", "Foo.jpg"}},
		{"c++", nil, nil}, // invalid regex falls back to substring
		{"*.zzz", nil, nil},
		{"zzz", nil, nil},
	}
	for _, c := range cases {
		gd, gf := filterEntries(dirs, files, c.pat)
		if !slices.Equal(entryNames(gd), c.wantD) || !slices.Equal(entryNames(gf), c.wantF) {
			t.Errorf("filterEntries(%q) = %v + %v, want %v + %v", c.pat, entryNames(gd), entryNames(gf), c.wantD, c.wantF)
		}
	}
}
