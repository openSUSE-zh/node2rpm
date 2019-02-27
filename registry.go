package main

import (
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

func convertInterfaceMap(m map[string]interface{}) map[string]string {
	res := map[string]string{}
	for k, v := range m {
		res[k] = v.(string)
	}
	return res
}

func joinMap(m1, m2 map[string]string) map[string]string {
	for ia, va := range m1 {
		if _, ok := m2[ia]; ok {
			m2[ia] += " " + va
			continue
		}
		m2[ia] = va
	}
	return m2
}

// RegistryQuery query registry to get a Package
func RegistryQuery(uri, ver string, dev bool, parents []string) Package {
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

	js, e := simplejson.NewJson(body)
	errChk(e)

	pkg := Package{}
	pkg.Name = js.Get("_id").MustString()

	if ver == "latest" {
		ver = js.Get("dist-tags").Get("latest").MustString()
	} else {
		pkg.Name = pkg.Name + "@" + ver
	}

	pkg.Version = ver
	//FIXME: license
	pkg.License = js.Get("license").MustString()
	if len(parents) == 0 {
		pkg.Parent = []string{}
	} else {
		pkg.Parent = parents
	}
	m, _ := js.Get("versions").Get(ver).Get("dependencies").Map()
	pkg.Dependencies = convertInterfaceMap(m)
	m, _ = js.Get("versions").Get(ver).Get("devDependencies").Map()
	pkg.DevDependencies = convertInterfaceMap(m)
	pkg.Tarball = js.Get("versions").Get(ver).Get("dist").Get("tarball").MustString()
	pkg.Child = joinMap(pkg.Dependencies, pkg.DevDependencies)
	fmt.Println(pkg)
	return pkg
}
