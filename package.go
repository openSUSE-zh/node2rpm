package main

import (
	"fmt"
	"github.com/Masterminds/semver"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type Parent struct {
	Name     string
	Brothers []string
}

type Parents []Parent

func (p Parents) Contains(s string) bool {
	for i := 0; i < len(p)-1; i++ {
		if p[i].Name == s {
			return true
		}
		for _, v := range p[i].Brothers {
			if v == s {
				return true
			}
		}
	}
	return false
}

func (p Parents) Inspect() string {
	s := p[0].Name
	for _, v := range p[0].Brothers {
		s += "\t" + v
	}
	s += "\n"

	idx := 1

	for i := 1; i < len(p)-1; i++ {
		s += strings.Repeat("\t", idx) + "|\n"
		s += strings.Repeat("\t", idx) + p[i].Name
		for _, v := range p[i].Brothers {
			s += "\t" + v
		}
		s += "\n"
		idx += 1
	}
	return s
}

// Dependency Tree
type Tree map[string]*Node

func (t Tree) Append(k string, v *Node, parents Parents) {
	tv := reflect.ValueOf(t)
	// i starts from 1 because first level of dependency is flattened to root tree
	// skip the last one because it's the one to be inserted.
	for i := 1; i < len(parents)-1; i++ {
		name := reflect.ValueOf(parents[i].Name)
		if tv.Kind() == reflect.Map {
			tv = tv.MapIndex(name)
		}
		if tv.Kind() == reflect.Ptr {
			tv = tv.Elem()
		}
		if tv.Kind() == reflect.Struct {
			if tv.FieldByName("Child").IsNil() {
				mapType := reflect.MapOf(reflect.TypeOf(k), reflect.TypeOf(v))
				tv.FieldByName("Child").Set(reflect.MakeMapWithSize(mapType, 0))
			}
			tv = tv.FieldByName("Child")
		}
	}
	tv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
}

func (t Tree) Inspect() string {
	
}

// Node of a dependency tree
type Node struct {
	License interface{} // SPDX License string or License Object
	Tarball string
	Child   Tree
}

// Package fetched from registry
type Package struct {
	Name     string
	Versions []string
	License  *simplejson.Json // SPDX License string or license Object
	Json     *simplejson.Json
}

// BuildDependencyTree build a dependency tree
func BuildDependencyTree(uri, ver string, tree Tree, parents Parents) {
	node := Node{}
	pkg := RegistryQuery(uri)

	// assign values to initialize the loop
	if ver == "latest" {
		ver = pkg.Versions[0]
	}

	if len(parents) == 0 {
		parents = append(parents, Parent{pkg.Name + "@" + ver, []string{}})
	}
	// end

	node.License = pkg.License
	node.Tarball = pkg.Json.Get(ver).Get("dist").Get("tarball").MustString()

	// root and the first level dependency
	if len(parents) < 3 {
		tree[pkg.Name+"@"+ver] = &node
	} else {
		// if parents already has this dependency, don't append
		if parents.Contains(pkg.Name + "@" + ver) {
			log.Printf("%s, version %s, has been provides via one of its parent, skiped.", pkg.Name, ver)
			fmt.Printf(parents.Inspect())
		} else {
			tree.Append(pkg.Name+"@"+ver, &node, parents)
		}
	}

	// calculate Child
	dependencies, _ := pkg.Json.Get(ver).Get("dependencies").Map()

	// calculate next parent
	sameDeepth := []string{}

	for k, constriant := range dependencies {
		childPkg := RegistryQuery(k)
		c, _ := constriant.(string)
		version := calculateSemver(childPkg.Versions, c)
		if len(version) == 0 {
			log.Fatalf("%s: no suitable version found for %s in %v.", k, constriant, childPkg.Versions)
		}
		sameDeepth = append(sameDeepth, k+"@"+version)
	}

	for i, k := range sameDeepth {
		left := []string{}
		for j, v := range sameDeepth {
			if i != j {
				left = append(left, v)
			}
		}
		parent := Parent{k, left}
		newParents := parents
		newParents = append(newParents, parent)
		a := strings.Split(k, "@")
		BuildDependencyTree(a[0], a[1], tree, newParents)
	}

}

func calculateSemver(versions []string, constriant string) string {
	c, e := semver.NewConstraint(constriant)
	errChk(e)

	for _, v := range versions {
		sv, e := semver.NewVersion(v)
		errChk(e)
		// always return the latest matched semver
		if c.Check(sv) {
			return v
		}
	}

	return ""
}

// RegistryQuery query registry to get informations of a Package
func RegistryQuery(uri string) Package {
	registry := "https://registry.npmjs.org/"
	pkgName := uri
	if strings.HasPrefix(uri, "http") {
		pkgName = filepath.Base(uri)
	}
	uri = registry + pkgName

	resp, e := http.Get(uri)
	errChk(e)
	defer resp.Body.Close()

	body := []byte{}

	if resp.StatusCode == http.StatusOK {
		body, e = ioutil.ReadAll(resp.Body)
		errChk(e)
	}

	if len(body) == 0 {
		log.Fatalf("Empty response body. Check whether your specified package exists: %s", uri)
	}

	js, e := simplejson.NewJson(body)
	errChk(e)

	pkg := Package{}
	pkg.Name = js.Get("_id").MustString()
	pkg.Json = js.Get("versions")
	versions, _ := pkg.Json.Map()
	pkg.Versions = getReverseSortedMapKeys(versions)

	//FIXME: license
	pkg.License = js.Get("license")

	return pkg
}

func getReverseSortedMapKeys(versions map[string]interface{}) []string {
	keys := []string{}
	for k := range versions {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	return keys
}
