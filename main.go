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

	temp := NewTempData()
	spec := Specfile{pkg, false, []byte{}, wd}

	if bundle {
		if len(exclude) > 0 {
			temp.Exclusion = parseExcludeString(exclude)
			log.Println("These packages are set to be excluded:")
			fmt.Println(temp.Exclusion.Inspect())
		} else {
			log.Println("No package to exclude, skipped.")
		}

		tree := Tree{}
		parentTree := ParentTree{}
		BuildDependencyTree(pkg, ver, tree, parentTree, Parents{}, temp)
		log.Printf("%s %s tree has been built:\n", pkg, ver)
		fmt.Println(tree.Inspect(0))
		tree.ToJson()
	} else {
		pkg1 := RegistryQuery(pkg, temp.ResponseCache)
		if ver == "latest" {
			ver = pkg1.Versions[0].String()
		}
		temp.Licenses.Append(pkg1.License)
		temp.Tarballs.Append(pkg1.Json.Get(ver).Get("dist").Get("tarball").MustString())
	}

	temp.Tarballs.ToService(wd)
	spec.Save(temp.Licenses.String(), temp.Tarballs.String())

	//FillSpecfile()
	log.Printf("Congrats! Module %s has been created/updated.", pkg)
}
