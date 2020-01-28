package main

import "testing"

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
