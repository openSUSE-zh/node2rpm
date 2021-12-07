package main

import (
	"reflect"
	"strings"

	"github.com/bitly/go-simplejson"
)

// Licenses holds all unique licenses of the tree
type Licenses map[string]struct{}

// Append appends new license to Licenses
func (licenses Licenses) Append(license string) {
	if _, ok := licenses[license]; !ok {
		licenses[license] = struct{}{}
	}
}

// String convert license map to RPM License string
func (licenses Licenses) String() string {
	m := map[string]struct{}{}
	for license := range licenses {
		if license == "Unlicense" {
			continue
		}
		if strings.Contains(license, " OR ") {
			for _, s := range strings.Split(license, " OR ") {
				if _, ok := m[s]; !ok {
					m[s] = struct{}{}
				}
			}
		} else {
			if _, ok := m[license]; !ok {
				m[license] = struct{}{}
			}
		}
	}
	keys := reflect.ValueOf(m).MapKeys()
	strKeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strKeys[i] = keys[i].String()
	}
	return strings.Join(strKeys, " AND ")
}

// getLicense parse license for package
// three kinds of license expression nowadays:
// 1. String {"license": "MIT"}
// 2. Array {"licenses": [{"type": "MIT", "url": "blabla"}, {"type": "Apache-2.0", "url":"daladala"}]}
// 3. Map {"license": {"type": "MIT", "url": "blabla"}}
// Both 2 and 3 are now deprecated but still in use.
func getLicense(js *simplejson.Json) string {
	j := js.Get("license")
	r := strings.NewReplacer("(", "", ")", "")

	s, e := j.String()
	if e == nil {
		if !strings.Contains(s, " OR ") {
			s = strings.Replace(s, " ", "-", -1)
		}
		return r.Replace(s)
	}

	m, e := j.Map()
	if e == nil {
		s, _ = m["type"].(string)
		return s
	}

	// the only way to check nil value for simplejson
	if reflect.ValueOf(j).Elem().Field(0).IsNil() {
		jv := js.Get("licenses").MustArray()
		a := []string{}
		for _, v := range jv {
			m := reflect.ValueOf(v).MapIndex(reflect.ValueOf("type")).Interface()
			s, _ = m.(string)
			a = append(a, s)
		}
		return strings.Join(a, " OR ")
	}

	return ""
}
