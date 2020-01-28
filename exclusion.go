package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

// Exclusion packages to be excluded. Useful for splitting a big bundle to several small bundles.
type Exclusion map[string]string

// Inspect debug output of Exclusion
func (e Exclusion) Inspect() string {
	s := "=== Packages to be excluded ===\n"
	s += "|\tPackage    |    Version    |\n"
	for k, v := range e {
		s += fmt.Sprintf("|\t%s    |    %s    |\n", k, v)
	}
	s += "=== END ==="
	return s
}

// Contains if a package with specified version locates in the Exclusion
func (e Exclusion) Contains(k string, v *semver.Version) bool {
	for m, n := range e {
		if k != m {
			continue
		}

		c, e := semver.NewConstraint(n)
		if e != nil {
			log.Fatalf("Could not initialize a new semver constriant froom %s", n)
		}

		if c.Check(v) {
			return true
		}
	}
	return false
}

func parseExcludeString(s string) Exclusion {
	e := Exclusion{}
	for _, v := range strings.Split(s, ",") {
		pkg, ver := parsePackageWithExplicitVersion(v)
		if len(ver) == 0 {
			e[pkg] = ">= 0.0.0"
		} else {
			re := regexp.MustCompile(`^\d`)
			if re.MatchString(ver) {
				e[pkg] = "= " + ver
			} else {
				// You can pass your own semver constriant
				e[pkg] = ver
			}
		}
	}
	return e
}

func parsePackageWithExplicitVersion(s string) (string, string) {
	a := strings.Split(s, ":")
	if len(a) < 2 {
		return a[0], ""
	}
	return a[0], a[1]
}
