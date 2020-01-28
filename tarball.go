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
