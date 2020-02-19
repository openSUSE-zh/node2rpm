package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

func parseService(wd string) map[string]struct{} {
	re := regexp.MustCompile(`path">(.*?)<`)
	tgz := make(map[string]struct{})
	service := filepath.Join(wd, "_service")
	if _, err := os.Stat(service); !os.IsNotExist(err) {
		f, _ := os.Open(service)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			t := scanner.Text()
			if re.MatchString(t) {
				t = re.FindStringSubmatch(t)[1]
				tgz[filepath.Base(t)] = struct{}{}
			}
		}
		f.Close()
	}
	return tgz
}
