package main

import (
	"encoding/json"
	"github.com/Masterminds/semver"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// Parent Parent contains the name of the direct parent and the direct parent's counterparts as brothers
type Parent struct {
	Name     string
	Brothers map[string]struct{}
}

// Parents a group of Parent, its indexing is important. smaller indexing means higher level of parent
type Parents []Parent

// Contains if a package is provided by one of its direct parents or direct parents' brothers
func (p Parents) Contains(s string) bool {
	for i := 0; i < len(p)-1; i++ {
		if p[i].Name == s {
			return true
		}
		if _, ok := p[i].Brothers[s]; ok {
			return true
		}
	}
	return false
}

// DirectParents the direct parents of a package
func (p Parents) DirectParents() []string {
	a := []string{}
	for _, v := range p {
		a = append(a, v.Name)
	}
	return a
}

// dedupeParents find intersection of two Parents
func dedupeParents(o, n Parents, t Tree) Parents {
	var low, high Parents
	if len(o)-len(n) >= 0 {
		low = n
		high = o
	} else {
		low = o
		high = n
	}
	idx := 0
	// "-2" because the last element is always different
	for i := len(low) - 2; i >= 0; i-- {
		if low[i].Name == high[i].Name {
			idx = i
			break
		}
	}
	// the last element' brothers should be the new direct parent's existing dependencies.
	brothers := map[string]struct{}{}
	for k := range t.FindChild(idx, low) {
		brothers[k] = struct{}{}
	}
	p := make([]Parent, idx+1, idx+1)
	copy(p, low)
	return append(p, Parent{low[len(low)-1].Name, brothers})
}

// ParentTree a place holding all nodes in the tree now with its parents
// used to compute unified parents for the unified package, or one package
// with a specified version may occur everywhere in the tree. (Dedupe)
type ParentTree map[string]Parents

// Tree Dependency Tree
type Tree map[string]*Node

// LoopFunc function to process struct in loop
type LoopFunc interface {
	Process(reflect.Value) reflect.Value
}

// AppendFunc the LoopFunc in Append method
type AppendFunc struct {
	Key   string
	Value *Node
}

// Process need to intialize the map for Append
func (fn AppendFunc) Process(tv reflect.Value) reflect.Value {
	if tv.FieldByName("Child").IsNil() {
		mapType := reflect.MapOf(reflect.TypeOf(fn.Key), reflect.TypeOf(fn.Value))
		tv.FieldByName("Child").Set(reflect.MakeMapWithSize(mapType, 0))
	}
	return tv
}

// DummyFunc the "do nothing" LoopFunc
type DummyFunc struct{}

// Process Dummy
func (fn DummyFunc) Process(tv reflect.Value) reflect.Value {
	return tv
}

// Loop loop through the tree to locate the element
func (t Tree) Loop(p Parents, fn LoopFunc) reflect.Value {
	tv := reflect.ValueOf(t)
	// skip the last element since it's the one to be processed
	for i := 0; i < len(p)-1; i++ {
		name := reflect.ValueOf(p[i].Name)
		if tv.Kind() == reflect.Map {
			tv = tv.MapIndex(name)
		}
		if tv.Kind() == reflect.Ptr {
			tv = tv.Elem()
		}
		if tv.Kind() == reflect.Struct {
			tv = fn.Process(tv)
			tv = tv.FieldByName("Child")
		}
	}
	return tv
}

// Append append an element to the tree
func (t Tree) Append(k string, v *Node, p Parents) {
	fn := AppendFunc{k, v}
	tv := t.Loop(p, fn)
	tv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
}

// Delete delete an element from the tree
func (t Tree) Delete(k string, p Parents) {
	fn := DummyFunc{}
	tv := t.Loop(p, fn)
	tv.SetMapIndex(reflect.ValueOf(k), reflect.Value{})
}

// FindChild find the child tree of the idx element of the parents
func (t Tree) FindChild(idx int, p Parents) Tree {
	tree := t
	for i := 0; i <= idx; i++ {
		tree = tree[p[i].Name].Child
	}
	return tree
}

// FindDependencies find dependencies of a node in the current tree
func (t Tree) FindDependencies(k string, p Parents) []reflect.Value {
	fn := DummyFunc{}
	tv := t.Loop(p, fn).MapIndex(reflect.ValueOf(k)).Elem().FieldByName("Child")
	keys := tv.MapKeys()
	if len(keys) > 0 {
		for _, v := range keys {
			nk := v.String()
			np := make([]Parent, len(p), cap(p))
			copy(np, p)
			np = append(np, Parent{nk, map[string]struct{}{}})
			keys = append(keys, t.FindDependencies(nk, np)...)
		}
	}
	return keys
}

// Inspect print the tree
func (t Tree) Inspect(idx int) string {
	s := ""
	for k, v := range t {
		s += strings.Repeat("\t", idx) + "|\n"
		s += strings.Repeat("\t", idx) + k + "\n"
		s += v.Child.Inspect(idx + 1)
	}
	return s
}

// ToJson write dependency tree to file
func (t Tree) ToJson(file string) {
	b, e := json.MarshalIndent(t, "", "\t")
	if e != nil {
		log.Fatalf("Can not convert the dependency tree to json: %s", e)
	}
	if !strings.HasSuffix(file, ".json") {
		file = file + ".json"
	}
	ioutil.WriteFile(file, b, 0644)
}

// Node the node structure of the tree
type Node struct {
	Child Tree `json:"child,omitempty"`
}

// Package package informations fetched from registry
type Package struct {
	Name     string
	Versions []*semver.Version
	License  string
	Json     *simplejson.Json
}

// Licenses holds all unique licenses of the tree
type Licenses map[string]struct{}

// Append appends new license to Licenses
func (l Licenses) Append(k string) {
	if _, ok := l[k]; !ok {
		l[k] = struct{}{}
	}
}

// Tarballs holds download uri of the module and its dependencies
type Tarballs map[string]struct{}

// Append appends new tarball to Tarballs
func (tb Tarballs) Append(uri string) {
	if _, ok := tb[uri]; !ok {
		tb[uri] = struct{}{}
	}
}

// ToService convert tarball map to _service
func (tb Tarballs) ToService(wd string) {
	s := "<services>\n"
	for k := range tb {
		s += "\t<service name=\"download_url\">\n"
		u, e := url.Parse(k)
		if e != nil {
			log.Fatalf("Can not parse download_url %s: %s", k, e)
		}
		s += "\t\t<param name=\"protocol\">" + u.Scheme + "</param>\n"
		s += "\t\t<param name=\"host\">" + u.Host + "</param>\n"
		s += "\t\t<param name=\"path\">" + u.Path + "</param>\n"
		s += "\t</service>\n"
	}
	s += "</services>\n"
	ioutil.WriteFile(filepath.Join(wd, "_service"), []byte(s), 0644)
}

// BuildDependencyTree build a dependency tree
func BuildDependencyTree(uri, ver string, cache RespCache, tree Tree, pt ParentTree, parents Parents, exclusion Exclusion, licenses Licenses, tarballs Tarballs) {
	node := Node{}
	pkg := RegistryQuery(uri, cache)
	ahead := true

	// assign values to initialize the loop
	if ver == "latest" {
		ver = pkg.Versions[0].String()
	}

	if len(parents) == 0 {
		parents = append(parents, Parent{pkg.Name + ":" + ver, map[string]struct{}{}})
	}
	// end

	licenses.Append(pkg.License)
	tarballs.Append(pkg.Json.Get(ver).Get("dist").Get("tarball").MustString())

	if len(parents) < 1 {
		// root
		tree[pkg.Name+":"+ver] = &node
		pt[pkg.Name+":"+ver] = parents
	} else {
		// if parents already has this dependency, don't append
		if parents.Contains(pkg.Name + ":" + ver) {
			log.Printf("%s, version %s, has been provided via one of its parents, skiped.", pkg.Name, ver)
			ahead = false
		} else {
			if ptParents, ok := pt[pkg.Name+":"+ver]; ok {
				log.Printf("%s, version %s, has been in the dependency tree but is not one of the new one's direct parents nor direct parents' counterparts, npm can not find it. try merging the old and the new to a place both can be found by their dependents.", pkg.Name, ver)
				log.Println("Computing an unified parent")
				parents = dedupeParents(ptParents, parents, tree)
				if reflect.DeepEqual(parents.DirectParents(), ptParents.DirectParents()) {
					log.Printf("Computed parent is exactly the same as the old parent, skipped")
					ahead = false
				} else {
					log.Println("Deleting existing old one from tree")
					// delete all dependencies of the deleted item from ParentTree as well
					d := tree.FindDependencies(pkg.Name+":"+ver, ptParents)
					tree.Delete(pkg.Name+":"+ver, ptParents)
					delete(pt, pkg.Name+":"+ver)
					for _, v := range d {
						delete(pt, v.String())
					}
					tree.Append(pkg.Name+":"+ver, &node, parents)
					pt[pkg.Name+":"+ver] = parents
				}
			} else {
				tree.Append(pkg.Name+":"+ver, &node, parents)
				pt[pkg.Name+":"+ver] = parents
			}
		}
	}

	// calculate Child
	if ahead {
		dependencies := getDependencies(pkg.Json.Get(ver).Get("dependencies"), cache, exclusion)
		if len(dependencies) > 0 {
			for i, k := range dependencies {
				left := map[string]struct{}{}
				for j, v := range dependencies {
					if i != j {
						left[v] = struct{}{}
					}
				}
				np := make([]Parent, len(parents), cap(parents))
				copy(np, parents)
				np = append(np, Parent{k, left})
				a := strings.Split(k, ":")
				BuildDependencyTree(a[0], a[1], cache, tree, pt, np, exclusion, licenses, tarballs)
			}
		}
	}
	// Child end
}

func getSemver(versions []*semver.Version, constriant string) *semver.Version {
	c, e := semver.NewConstraint(constriant)
	if e != nil {
		log.Fatalf("Could not initialize a new semver constriant from %s", constriant)
	}

	for _, v := range versions {
		// always return the latest matched semver
		if c.Check(v) {
			return v
		}
	}

	return &semver.Version{}
}

func getDependencies(js *simplejson.Json, cache RespCache, exclusion Exclusion) []string {
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
		childPkg := RegistryQuery(k, cache)
		c, _ := constriant.(string)
		version := getSemver(childPkg.Versions, c)
		if len(version.String()) == 0 {
			log.Fatalf("%s: no suitable version found for %s in %v.", k, constriant, childPkg.Versions)
		}
		if exclusion.Contains(k, version) {
			log.Printf("%s version %s matched one of the packages known to be excluded, skipped.", k, version.String())
		} else {
			dependencies = append(dependencies, k+":"+version.String())
		}
	}

	return dependencies
}

// RespCache cache the http response to optimize query time
type RespCache map[string][]byte

// RegistryQuery query registry to get informations of a Package
func RegistryQuery(uri string, cache RespCache) Package {
	formatURI(&uri)
	body := getHttpBody(uri, cache)

	js, e := simplejson.NewJson(body)
	if e != nil {
		log.Fatalf("Cannot parse to json %s", body)
	}

	pkg := Package{}
	pkg.Name = js.Get("_id").MustString()
	pkg.Json = js.Get("versions")
	versions, _ := pkg.Json.Map()
	pkg.Versions = getReverseSorted(versions)
	pkg.License = getLicense(js)

	return pkg
}

// formatUri standardlize registry uri in place
func formatURI(uri *string) {
	registry := "https://registry.npmjs.org/"
	if strings.HasPrefix(*uri, "http") {
		*uri = filepath.Base(*uri)
	}
	if strings.Contains(*uri, "@") {
		log.Printf("scoped package found %v", *uri)
		*uri = strings.Replace(*uri, "@", "%40", -1)
		*uri = strings.Replace(*uri, "/", "%2F", -1)
	}
	*uri = registry + *uri
}

func getHttpBody(uri string, cache RespCache) []byte {
	// use cache first
	if body, ok := cache[uri]; ok {
		// all error handling has been done in the first time
		return body
	} else {
		resp, e := http.Get(uri)
		if e != nil {
			log.Fatalf("Can't get http response from %s", uri)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, e := ioutil.ReadAll(resp.Body)
			if e != nil {
				log.Fatalf("Can't read http body %v", resp.Body)
			}
			if len(body) == 0 {
				log.Fatalf("Empty response body. Check whether your specified package exists: %s", uri)
			}
			// dump body
			cache[uri] = body
			return body
		} else {
			log.Fatalf("statuscode of %s request is not 200 but %d", uri, resp.StatusCode)
		}
	}
	return []byte{}
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

// getReverseSorted reverse sort the available versions because newer
// version tends to be used frequently. save a lot of match work
func getReverseSorted(versions map[string]interface{}) []*semver.Version {
	ver := []*semver.Version{}
	for v := range versions {
		sv, e := semver.NewVersion(v)
		if e != nil {
			log.Fatalf("Can not build semver from %s.", v)
		}
		ver = append(ver, sv)
	}
	sort.Sort(sort.Reverse(semver.Collection(ver)))
	return ver
}
