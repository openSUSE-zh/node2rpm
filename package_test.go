package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	simplejson "github.com/bitly/go-simplejson"
	//"reflect"
	"testing"
)

func Test_getLicense(t *testing.T) {
	cases := map[string]string{"map": "{ \"license\" :\n\t{ \"type\" : \"ISC\",\n\t  \"url\" : \"https://opensource.org/licenses/ISC\"\n\t}\n}\n",
		"string": "{ \"license\": \"ISC\" }\n",
		"array":  "{ \"licenses\" : [{\n\t   \"type\": \"MIT\",\n\t   \"url\": \"https://www.opensource.org/licenses/mit-license.php\"\n\t},\n\t{\n\t   \"type\": \"Apache-2.0\",\n\t   \"url\": \"https://opensource.org/licenses/apache2.0.php\"\n\t}\n  ]\n}\n"}
	answers := map[string]string{"map": "ISC", "string": "ISC", "array": "MIT OR Apache-2.0"}

	for k, v := range cases {
		b := []byte(v)
		js, _ := simplejson.NewJson(b)
		license := getLicense(js)
		if license == answers[k] {
			t.Logf("Test getLicense() with type %s succeed", k)
		} else {
			t.Errorf("Test getLicense() with type %s failed, expected %s, got %s", k, answers[k], license)
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

func Test_formatURI(t *testing.T) {
	cases := []string{"https://registry.npmjs.org/punycode", "punycode", "@type/node"}
	registry := "https://registry.npmjs.org/"
	answers := []string{"punycode", "punycode", "%40type%2Fnode"}
	for i, v := range cases {
		uri := v
		formatURI(&uri)
		if uri == registry+answers[i] {
			t.Logf("Test formatURI() succeed, expected %s, got %s", registry+answers[i], uri)
		} else {
			t.Errorf("Test formatURI() failed, expected %s, got %s", registry+answers[i], uri)
		}
	}
}

/*func Test_dedupeParents(t *testing.T) {
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
}*/

func Test_rewriteConstriantWithExplicitComma(t *testing.T) {
	cases := []string{">= 2.1.2 < 3", "~ 2.1.2", "^2.x || >= 2.1.2 < 3"}
	answers := []string{">= 2.1.2, < 3", "~ 2.1.2", "^2.x || >= 2.1.2, < 3"}
	for k, v := range cases {
		answer := rewriteConstriantWithExplicitComma(v)
		if answer == answers[k] {
			t.Logf("Test rewriteConstriantWithExplicitComma succeed, expected %s, got %s", answers[k], answer)
		} else {
			t.Errorf("Test rewriteConstriantWithExplicitComman failed, expected %s, got %s", answers[k], answer)
		}
	}
}
