package main

import "testing"

func Test_parsePackageWithExplicitVersion(t *testing.T) {
	s := "rimraf@1.0.0"
	pkg, ver := parsePackageWithExplicitVersion(s)
	if pkg == "rimraf" {
		if ver == "1.0.0" {
			t.Log("Test passed")
		} else {
			t.Errorf("Parsed package version is wrong: %s, should be '1.0.0'", ver)
		}
	} else {
		t.Errorf("Parsed package name is wrong: %s, should be 'rimraf'", pkg)
	}
}
