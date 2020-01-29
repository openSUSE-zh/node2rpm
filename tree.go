package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	simplejson "github.com/bitly/go-simplejson"
)

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
func (t Tree) ToJson() {
	file := strings.Replace(reflect.ValueOf(t).MapKeys()[0].String(), ":", "-", -1)
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

// BuildDependencyTree build a dependency tree
func BuildDependencyTree(uri string, ver *string, tree Tree, pt ParentTree, parents Parents, temp TempData) {
	node := Node{}
	pkg := RegistryQuery(uri, temp.ResponseCache)
	ahead := true

	// assign values to initialize the loop
	if *ver == "latest" {
		*ver = pkg.Versions[0].String()
	}

	if len(parents) == 0 {
		parents = append(parents, Parent{pkg.Name + ":" + *ver, map[string]struct{}{}})
	}
	// end

	temp.Licenses.Append(pkg.License)
	temp.Tarballs.Append(pkg.Json.Get(*ver).Get("dist").Get("tarball").MustString())

	if len(parents) < 1 {
		// root
		tree[pkg.Name+":"+*ver] = &node
		pt[pkg.Name+":"+*ver] = parents
	} else {
		// if parents already has this dependency, don't append
		if parents.Contains(pkg.Name + ":" + *ver) {
			log.Printf("%s, version %s, has been provided via one of its parents, skiped.", pkg.Name, *ver)
			ahead = false
		} else {
			if ptParents, ok := pt[pkg.Name+":"+*ver]; ok {
				log.Printf("%s, version %s, has been in the dependency tree but is not one of the new one's direct parents nor direct parents' counterparts, npm can not find it. try merging the old and the new to a place both can be found by their dependents.", pkg.Name, *ver)
				log.Println("Computing an unified parent")
				parents = dedupeParents(ptParents, parents, tree)
				if reflect.DeepEqual(parents.DirectParents(), ptParents.DirectParents()) {
					log.Printf("Computed parent is exactly the same as the old parent, skipped")
					ahead = false
				} else {
					log.Println("Deleting existing old one from tree")
					// delete all dependencies of the deleted item from ParentTree as well
					d := tree.FindDependencies(pkg.Name+":"+*ver, ptParents)
					tree.Delete(pkg.Name+":"+*ver, ptParents)
					delete(pt, pkg.Name+":"+*ver)
					for _, v := range d {
						delete(pt, v.String())
					}
					tree.Append(pkg.Name+":"+*ver, &node, parents)
					pt[pkg.Name+":"+*ver] = parents
				}
			} else {
				tree.Append(pkg.Name+":"+*ver, &node, parents)
				pt[pkg.Name+":"+*ver] = parents
			}
		}
	}

	// calculate Child
	if ahead {
		dependencies := getDependencies(pkg.Json.Get(*ver).Get("dependencies"), temp.ResponseCache, temp.Exclusion)
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
				s := a[1]
				BuildDependencyTree(a[0], &s, tree, pt, np, temp)
			}
		}
	}
	// Child end
}

// rewriteConstriantWithExplicitComma node semver uses implicit "AND", eg ">= 2.1.2 < 3"
// but masterminds semver requires explicit comma as "AND", eg ">= 2.1.2, < 3"
func rewriteConstriantWithExplicitComma(s string) string {
	// keep the "or" condition
	strs := strings.Split(s, "||")
	// identify how many pairs. ">= 2.1.2" is one pair, the operator
	// needs to be in front of the version number.
	re := regexp.MustCompile(`[<=>^~]+(\s+)?\d+(\.[^\s]+)?`)
	for k, v := range strs {
		m := re.FindAllStringSubmatch(v, -1)
		if len(m) > 0 {
			for i := 0; i < len(m)-1; i++ {
				v = strings.Replace(v, m[i][0], m[i][0]+",", 1)
			}
			strs[k] = v
		}
	}
	// restore the "or" condition
	s = strings.Join(strs, "||")
	return s
}

func getSemver(versions []*semver.Version, constriant string) *semver.Version {
	c, e := semver.NewConstraint(rewriteConstriantWithExplicitComma(constriant))
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

func getDependencies(js *simplejson.Json, cache ResponseCache, exclusion Exclusion) []string {
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
