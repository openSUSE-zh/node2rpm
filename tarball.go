package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
)

// Tarballs holds download uri of the module and its dependencies
type Tarballs map[string]struct{}

// Append appends new tarball to Tarballs
func (tb Tarballs) Append(uri string) {
	if _, ok := tb[uri]; !ok {
		tb[uri] = struct{}{}
	}
}

func (tb Tarballs) diff(m map[string]struct{}, wd string) {
	s := "#!/bin/bash\n"
	for k := range tb {
		tgz := filepath.Base(k)
		if _, ok := m[tgz]; !ok {
			log.Printf("%s should be removed\n", tgz)
			s += "osc delete " +tgz+"\n"
		}
	}
	ioutil.WriteFile(filepath.Join(wd, "remove.sh"), []byte(s), 0755)
}

// ToService convert tarball map to _service
func (tb Tarballs) ToService(wd string) {
	s := "<services>\n"
	for k := range tb {
		s += "\t<service name=\"download_url\" mode=\"localonly\">\n"
		u, e := url.Parse(k)
		if e != nil {
			log.Fatalf("Can not parse download_url %s: %s", k, e)
		}
		s += "\t\t<param name=\"protocol\">" + u.Scheme + "</param>\n"
		s += "\t\t<param name=\"host\">" + u.Host + "</param>\n"
		s += "\t\t<param name=\"path\">" + u.Path + "</param>\n"
		s += "\t</service>\n"
	}
	s += "</services>\n"

	m := parseService(wd)
	if m != nil {
		tb.diff(m, wd)
	}
	ioutil.WriteFile(filepath.Join(wd, "_service"), []byte(s), 0644)
}

// String convert tarball map to RPM Source string
func (tb Tarballs) String() string {
	idx := 0
	s := ""
	for k := range tb {
		s += "Source" + strconv.Itoa(idx) + ":\t" + k + "\n"
		idx += 1
	}
	return s
}
