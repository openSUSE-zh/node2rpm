package main

import (
	//"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Specfile struct {
	Name             string
	Templated        bool
	Raw              []byte
	WorkingDirectory string
}

func (s Specfile) Load() {
	spec := filepath.Join(s.WorkingDirectory, s.Name+".spec")
	var e error
	if _, e = os.Stat(spec); os.IsNotExist(e) {
		// specfile does not exist in current working directory, use template to initialize new
		s.Templated = true
		s.Raw, e = ioutil.ReadFile("node2rpm.template")
	} else {
		s.Templated = false
		s.Raw, e = ioutil.ReadFile(spec)
	}
	if e != nil {
		s.Raw = []byte{}
		log.Println("Can not find or read specfile template.")
	}
}

func (s Specfile) Save(licenses, sources string) {
	ioutil.WriteFile(filepath.Join(s.WorkingDirectory, s.Name+".spec"), s.Raw, 0644)
}
