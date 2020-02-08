package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	//"github.com/Masterminds/semver"
	simplejson "github.com/bitly/go-simplejson"
	semver "github.com/openSUSE-zh/node-semver"
)

// Package package informations fetched from registry
type Package struct {
	Name     string
	Versions semver.Collection
	License  string
	Json     *simplejson.Json
}

// ResponseCache cache the http response to optimize query time
type ResponseCache map[string][]byte

// RegistryQuery query registry to get informations of a Package
func RegistryQuery(uri string, cache ResponseCache) Package {
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

func getHttpBody(uri string, cache ResponseCache) []byte {
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

// getReverseSorted reverse sort the available versions because newer
// version tends to be used frequently. save a lot of match work
func getReverseSorted(versions map[string]interface{}) semver.Collection {
	ver := semver.Collection{}
	for v := range versions {
		sv := semver.NewSemver(v)
		//if e != nil {
		//	log.Fatalf("Can not build semver from %s.", v)
		//}
		ver = append(ver, sv)
	}
	sort.Sort(sort.Reverse(ver))
	return ver
}
