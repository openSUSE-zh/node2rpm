package main

type Package struct {
	Name            string
	Version         string
	Dependencies    map[string]string
	DevDependencies map[string]string
	License         string
	Tarball         string
	Parent          []string
	Child           map[string]string
}
