package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	var pkg, ver string
	var dev bool
	flag.StringVar(&pkg, "pkg", "", "the module needs to package.")
	flag.StringVar(&ver, "v", "latest", "the module's version.")
	flag.BoolVar(&dev, "dev", false, "whether to package devDependencies.")
	flag.Parse()

	if len(pkg) == 0 {
		log.Fatal("You must specify a module name to package.")
	}

	fmt.Println(RegistryQuery(pkg, ver, dev, []string{}))
}
