package main

import (
	simplejson "github.com/bitly/go-simplejson"
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
