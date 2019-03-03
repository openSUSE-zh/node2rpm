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
	// skip the last one because it's the one to be inserted.
	for i := 0; i < len(parents)-1; i++ {
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

func (t Tree) Inspect(idx int) string {
	s := ""
	for k, v := range t {
		s += strings.Repeat("\t", idx) + "|\n"
		s += strings.Repeat("\t", idx) + k + "\n"
		s += v.Child.Inspect(idx + 1)
	}
	return s
}

// Node of a dependency tree
type Node struct {
	License string
	Tarball string
	Parent  Parents
	Child   Tree
}

// Package fetched from registry
type Package struct {
	Name     string
	Versions []string
	License  string
	Json     *simplejson.Json
}

// BuildDependencyTree build a dependency tree
func BuildDependencyTree(uri, ver string, tree Tree, parents Parents, ex Exclusion) {
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

	// see explanations below
	node.Parent = parents[:len(parents)-1]
	node.License = pkg.License
	node.Tarball = pkg.Json.Get(ver).Get("dist").Get("tarball").MustString()

	if len(parents) < 1 {
		// root
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
	dependencies := getDependencies(pkg.Json.Get(ver).Get("dependencies"), ex)

	for i, k := range dependencies {
		left := []string{}
		for j, v := range dependencies {
			if i != j {
				left = append(left, v)
			}
		}
		newParents := append(parents, Parent{k, left})
		// next Parent end
		a := strings.Split(k, "@")
		BuildDependencyTree(a[0], a[1], tree, newParents, ex)
	}
	// Child end
}

func getSemver(versions []string, constriant string) string {
	c, e := semver.NewConstraint(constriant)
	if e != nil {
		log.Fatalf("Could not initialize a new semver constriant froom %s", constriant)
	}

	for _, v := range versions {
		sv, e := semver.NewVersion(v)
		if e != nil {
			log.Fatalf("Could not initialize a new semver version from %s", v)
		}

		// always return the latest matched semver
		if c.Check(sv) {
			return v
		}
	}

	return ""
}

func getDependencies(js *simplejson.Json, ex Exclusion) []string {
	upstreamDependencies, _ := js.Map()
	// calculate next parent, we need to append current dependencies as parents
	// for packages in the next loop here in this loop. because in the next loop,
	// we have no way to find the counterparts of the package's direct parent. eg:
	// Loop 1: A and B, B's dependencies is C, D.
	// Loop 2: C, C's dependency is A
	// C loop is triggered by B loop. B loop only knows B's dependencies C and D.
	// B doesn't know it's counterpart A. because such dependencies are only known
	// to B's parent. so C doesn't know A is in its up-level too.
	// So we append all dependencies of B's parent (including B itself) as the last
	// parent of B, eg:
	// Loop 1: A and B, B's dependencies is C, D. B's parent is [whatever, [A, B]]
	// Loop 2: C. C's dependency is A. C's parent is [whatever, [A, B], [C, D]].
	// Now C knows A. With a clever design (see Parent type), C also knows its direct
	// parent is B.
	// With this design. when calculating parents, we need to skip the last parent, eg:
	// Loop B, parent [whatever, [A, B]].
	// We need to skip [A, B], or our resolver will think B has already been in the tree.
	dependencies := []string{}

	for k, constriant := range upstreamDependencies {
		childPkg := RegistryQuery(k)
		c, _ := constriant.(string)
		version := getSemver(childPkg.Versions, c)
		if len(version) == 0 {
			log.Fatalf("%s: no suitable version found for %s in %v.", k, constriant, childPkg.Versions)
		}
		if ex.Contains(k, version) {
			log.Printf("%s version %s matched one of the packages known to be excluded, skipped.", k, version)
		} else {
			dependencies = append(dependencies, k+"@"+version)
		}
	}

	return dependencies
}

// RegistryQuery query registry to get informations of a Package
func RegistryQuery(uri string) Package {
	formatUri(&uri)
	body := getHttpBody(uri)

	js, e := simplejson.NewJson(body)
	errChk(e)

	pkg := Package{}
	pkg.Name = js.Get("_id").MustString()
	pkg.Json = js.Get("versions")
	versions, _ := pkg.Json.Map()
	pkg.Versions = getReverseSortedMapKeys(versions)
	pkg.License = getLicense(js)

	return pkg
}

// formatUri standardlize registry uri in place
func formatUri(uri *string) {
	registry := "https://registry.npmjs.org/"
	if strings.HasPrefix(*uri, "http") {
		*uri = filepath.Base(*uri)
	}
	*uri = registry + *uri
}

func getHttpBody(uri string) []byte {
	resp, e := http.Get(uri)
	if e != nil {
		log.Fatalf("Can't get http response from %s", uri)
	}
	defer resp.Body.Close()

	body := []byte{}

	if resp.StatusCode == http.StatusOK {
		body, e = ioutil.ReadAll(resp.Body)
		if e != nil {
			log.Fatalf("Can't read http body %v", resp.Body)
		}
		if len(body) == 0 {
			log.Fatalf("Empty response body. Check whether your specified package exists: %s", uri)
		}
	} else {
		log.Fatalf("statuscode is not 200 but %d", resp.StatusCode)
	}

	return body
}

// getLicense parse license for package
// three kinds of license expression nowadays:
// 1. String {"license": "MIT"}
// 2. Array {"licenses": [{"type": "MIT", "url": "blabla"}, {"type": "Apache-2.0", "url":"daladala"}]}
// 3. Map {"license": {"type": "MIT", "url": "blabla"}}
// Both 2 and 3 are now deprecated but still in use.
func getLicense(js *simplejson.Json) string {
	j := js.Get("license")

	s, e := j.String()
	if e == nil {
		return s
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

// getReverseSortedMapKeys reverse sort the available versions because newer
// version tends to be used frequently. save a lot of match work
func getReverseSortedMapKeys(versions map[string]interface{}) []string {
	keys := []string{}
	for k := range versions {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	return keys
}
