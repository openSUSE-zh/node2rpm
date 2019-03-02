package main

import (
	"flag"
	"log"
)

func main() {
	var pkg, ver string
	//var bundle bool
	flag.StringVar(&pkg, "pkg", "", "the module needs to package.")
	flag.StringVar(&ver, "ver", "latest", "the module's version.")
	//flag.BoolVar(&bundle, "bundle", true, "don't bundle dependencies.")
	flag.Parse()

	if len(pkg) == 0 {
		log.Fatal("You must specify a module name to package.")
	}

	tree := Tree{}
	BuildDependencyTree(pkg, ver, tree, Parents{})
	log.Println(tree["ajv@6.9.2"].Child["uri-js@4.2.2"].Child)

	//	if bundle {
	//GenerateJson()

	//	}

	//Download()

	//FillSpecfile()
	log.Printf("Congrats! Module %s has been created/updated.", pkg)
}
