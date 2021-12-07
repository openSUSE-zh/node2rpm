package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Specfile
type Specfile struct {
	Name             string
	Templated        bool
	Raw              []byte
	WorkingDirectory string
}

// NewSpecfile initialize a new Specfile structure
func NewSpecfile(name, wd, specTemplate string) Specfile {
	spec := filepath.Join(wd, name+".spec")
	templated := false

	if _, err := os.Stat(spec); os.IsNotExist(err) {
		templated = true
		spec = specTemplate
	}

	raw, err := ioutil.ReadFile(spec)
	if err != nil {
		log.Printf("Can not find or read specfile %s", filepath.Join(wd, name+".spec"))
	}

	return Specfile{name, templated, raw, wd}
}

func (s *Specfile) Fill(pkg, ver string, bundle bool, temp TempData) {
	raw := string(s.Raw)
	if s.Templated {
		raw = strings.Replace(raw, "<PACKAGE>", pkg, -1)
		raw = strings.Replace(raw, "<VERSION>", ver, -1)
		raw = strings.Replace(raw, "<SOURCE>", temp.Tarballs.String(), 1)
		raw = strings.Replace(raw, "<LICENSE>", temp.Licenses.String(), 1)
	} else {

	}
	s.Raw = []byte(raw)
}

func (s Specfile) Save() {
	ioutil.WriteFile(filepath.Join(s.WorkingDirectory, s.Name+".spec"), s.Raw, 0644)
}
