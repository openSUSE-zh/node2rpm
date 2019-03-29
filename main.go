package main

import (
	"flag"
	"fmt"
	"github.com/Masterminds/semver"
	"log"
	"regexp"
	"strings"
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

func main() {
	var pkg, ver, exclude string
	var bundle bool
	flag.StringVar(&pkg, "pkg", "", "the module needs to package.")
	flag.StringVar(&ver, "ver", "latest", "the module's version.")
	flag.BoolVar(&bundle, "bundle", true, "don't bundle dependencies.")
	flag.StringVar(&exclude, "exclude", "", "the module to be excluded, in 'rimraf:1.0.0,mkdirp:1.0.1' format.")
	flag.Parse()

	if len(pkg) == 0 {
		log.Fatal("You must specify a module name to package.")
	}

	if bundle {
		ex := Exclusion{}
		if len(exclude) > 0 {
			ex = parseExcludeString(exclude)
			log.Println("These packages are set to be excluded:")
			fmt.Println(ex.Inspect())
		} else {
			log.Println("No package to exclude, skipped.")
		}

		tree := Tree{}
		parentTree := ParentTree{}
		licenses := Licenses{}
		BuildDependencyTree(pkg, ver, tree, parentTree, Parents{}, ex, licenses)
		log.Printf("%s %s tree has been built:\n", pkg, ver)
		fmt.Println(tree.Inspect(1))
		fmt.Println(licenses)
	}

	//	if bundle {
	//GenerateJson()

	//	}

	//Download()

	//FillSpecfile()
	log.Printf("Congrats! Module %s has been created/updated.", pkg)
}
