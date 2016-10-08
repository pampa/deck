package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

func matchAny(p string, a []*regexp.Regexp) bool {
	for _, r := range a {
		if m := r.MatchString(p); m == true {
			return true
		}
	}
	return false
}

func printFiles(cap string, files []string) {
	if len(files) > 0 {
		fmt.Println(cap, ":\n")
		for _, f := range files {
			fmt.Println("\t", f)
		}
		fmt.Println()
	}
}

func getConfigFile(cFlag string) (string, error) {
	var cFiles []string

	if cFlag != "" {
		cFiles = append([]string{}, cFlag)
	} else {
		cFiles = append([]string{os.ExpandEnv("${HOME}/.deckrc"), "/etc/deckrc"})
	}

	var cFile string

	for _, f := range cFiles {
		log.Debug("try config file " + f)
		if finfo, err := os.Stat(f); err == nil {
			if !finfo.IsDir() {
				log.Debug("using " + f)
				cFile = f
				break
			}
		}
	}

	if cFile == "" {
		return cFile, errors.New("can't find config file")
	} else {
		return cFile, nil
	}
}
