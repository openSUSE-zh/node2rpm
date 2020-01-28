package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_ToService(t *testing.T) {
	tb := Tarballs{}
	tb["https://registry.npmjs.org/punycode/-/punycode-2.1.1.tgz"] = struct{}{}
	wd := "/tmp"
	tb.ToService(wd)
	f := filepath.Join(wd, "_service")
	dat, e := ioutil.ReadFile(f)
	if e != nil {
		t.Errorf("Test Tarballs.ToService() failed: can not read _service.")
	}

	answer := "<services>\n\t<service name=\"download_url\" mode=\"localonly\">\n\t\t<param name=\"protocol\">https</param>\n\t\t<param name=\"host\">registry.npmjs.org</param>\n\t\t<param name=\"path\">/punycode/-/punycode-2.1.1.tgz</param>\n\t</service>\n</services>\n"
	if string(dat) == answer {
		t.Log("Test Tarballs.ToService() passed")
	} else {
		t.Errorf("Test Tarballs.ToService() failed: expected\n %s\n, got\n %s", answer, string(dat))
	}
	os.Remove(f)
}
