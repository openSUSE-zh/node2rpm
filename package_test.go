package main

import (
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_getLicense(t *testing.T) {
	testCases := map[string]string{"map": "{ \"license\" :\n\t{ \"type\" : \"ISC\",\n\t  \"url\" : \"https://opensource.org/licenses/ISC\"\n\t}\n}\n",
		"string": "{ \"license\": \"ISC\" }\n",
		"array":  "{ \"licenses\" : [{\n\t   \"type\": \"MIT\",\n\t   \"url\": \"https://www.opensource.org/licenses/mit-license.php\"\n\t},\n\t{\n\t   \"type\": \"Apache-2.0\",\n\t   \"url\": \"https://opensource.org/licenses/apache2.0.php\"\n\t}\n  ]\n}\n"}
	testResults := map[string]string{"map": "ISC", "string": "ISC", "array": "MIT OR Apache-2.0"}

	for k, v := range testCases {
		b := []byte(v)
		js, _ := simplejson.NewJson(b)
		license := getLicense(js)
		if license == testResults[k] {
			t.Logf("getLicense(): %s test passed.", k)
		} else {
			t.Errorf("getLicense(): %s test failed.", k)
		}
	}
}

func Test_ToService(t *testing.T) {
	tb := Tarballs{}
	tb["https://registry.npmjs.org/punycode/-/punycode-2.1.1.tgz"] = struct{}{}
	wd := "/tmp"
	tb.ToService(wd)
	f := filepath.Join(wd, "_service")
	dat, e := ioutil.ReadFile(f)
	if e != nil {
		t.Errorf("Test failed: can not read _service.")
	}

	result := "<services>\n\t<service name=\"download_url\">\n\t\t<param name=\"protocol\">https</param>\n\t\t<param name=\"host\">registry.npmjs.org</param>\n\t\t<param name=\"path\">/punycode/-/punycode-2.1.1.tgz</param>\n\t</service>\n</services>\n"
	if result == string(dat) {
		t.Log("Test passed")
	} else {
		t.Errorf("Test failed: expected\n %s, got\n %s.", result, string(dat))
	}
	os.Remove(f)
}

func Test_dedupeParents(t *testing.T) {
	var r Parents
	brothers := map[string]struct{}{}
	r = append(r, Parent{"root", brothers})
	r = append(r, Parent{"rimraf@1.0.0", brothers})
	r = append(r, Parent{"wrappy@1.0.0", brothers})
	o := make(Parents, len(r))
	n := make(Parents, len(r))
	copy(o, r)
	copy(n, r)
	r = append(r, Parent{})
	o = append(o, Parent{"A", brothers})
	n = append(n, Parent{"B", brothers})
	n = append(n, Parent{"C", brothers})
	testResult := dedupeParents(o, n)
	if reflect.DeepEqual(testResult, r) {
		t.Log("Test passed")
	} else {
		t.Errorf("dedupeParents() failed with result %v, should be %v", testResult, r)
	}
}
