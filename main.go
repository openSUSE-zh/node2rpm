package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	currentWd, e := os.Getwd()
	if e != nil {
		log.Fatal(e)
	}

	var pkg, ver, exclude, wd string
	var bundle bool
	flag.StringVar(&pkg, "pkg", "", "the module needs to package.")
	flag.StringVar(&ver, "ver", "latest", "the module's version.")
	flag.BoolVar(&bundle, "bundle", true, "don't bundle dependencies.")
	flag.StringVar(&exclude, "exclude", "", "the module to be excluded, in 'rimraf:1.0.0,mkdirp:1.0.1' format.")
	flag.StringVar(&wd, "wd", currentWd, "the osc working directory")
	flag.Parse()

	if len(pkg) == 0 {
		log.Fatal("You must specify a module name to package.")
	}

	cache := RespCache{}
	spec := Specfile{pkg, false, []byte{}, wd}

	if bundle {
		exclusion := Exclusion{}
		if len(exclude) > 0 {
			exclusion = parseExcludeString(exclude)
			log.Println("These packages are set to be excluded:")
			fmt.Println(exclusion.Inspect())
		} else {
			log.Println("No package to exclude, skipped.")
		}

		tree := Tree{}
		parentTree := ParentTree{}
		licenses := Licenses{}
		tarballs := Tarballs{}
		BuildDependencyTree(pkg, ver, cache, tree, parentTree, Parents{}, exclusion, licenses, tarballs)
		log.Printf("%s %s tree has been built:\n", pkg, ver)
		fmt.Println(tree.Inspect(0))
		tree.ToJson()
		tarballs.ToService(wd)
		spec.Save(licenses.String(), tarballs.String())
	} else {

	}

	//FillSpecfile()
	log.Printf("Congrats! Module %s has been created/updated.", pkg)
}
