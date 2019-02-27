package main

import "log"

func errChk(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
