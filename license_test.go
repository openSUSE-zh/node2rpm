package main

import (
	"testing"

	simplejson "github.com/bitly/go-simplejson"
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
