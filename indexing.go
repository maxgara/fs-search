package main

import (
	"fmt"
	"os"
	"slices"
	"unicode"
)

var TARGET_FILETYPES = []string{"json", "html", "csv", "py", "text", "c", "lua", "cpp", "sh", "txt", "xml", "yaml"}

var allFiletypes map[string]bool

// recursively read filenames in directory rootPath, return target file names
func ReadWalk(rootpath string) []string {
	allFiletypes = make(map[string]bool)
	var targetFilenames []string
	rReadWalk(rootpath, &allFiletypes, &targetFilenames)
	return targetFilenames
}

func rReadWalk(rootPath string, allFiletypes *map[string]bool, targetFilenames *[]string) {
	files, err := os.ReadDir(rootPath)
	if err != nil {
		fmt.Printf("ReadWalk dir: %s: %v\n", rootPath, err)
	}
	for _, f := range files {
		nm := f.Name()
		full := fmt.Sprintf("%v/%v", rootPath, nm)
		if f.IsDir() {
			rReadWalk(full, allFiletypes, targetFilenames)
			continue
		}

		//save target filenames
		for i := len(nm) - 2; i >= 0; i-- {
			if nm[i] != '.' {
				continue
			}
			ft := nm[i+1:] // filetype eg: "csv"
			(*allFiletypes)[ft] = true
			if slices.Contains(TARGET_FILETYPES, ft) {
				*targetFilenames = append(*targetFilenames, full)
			}
			break
		}

	}
}

func fileWords(path string) []string {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("fileWords error reading %s: %v\n", path, err)
		return nil
	}
	var words []string
	var startIdx int
	var hasLetter bool
	for i, b := range string(raw) {
		//non-word char case (end word):
		if !unicode.IsLetter(b) && !unicode.IsNumber(b) {
			//skip words with 0 letters
			if hasLetter {
				words = append(words, string(raw)[startIdx:i])
				hasLetter = false
			}
			startIdx = -1
			continue
		}
		if unicode.IsLetter(b) {
			hasLetter = true
		}
		//start word, if not yet started
		if startIdx == -1 {
			startIdx = i
		}
	}
	return words
}

const HASH_NONCE = 13

func hash(data []byte) uint32 {
	var out uint32
	for _, b := range data {
		out *= HASH_NONCE
		out += uint32(b)
	}
	return out
}

type Dictionary struct {
	files []string
	data  []wloc
}
type wloc struct {
	key  uint32
	fidx int
}

func addFileKeys(path string, dx *Dictionary) {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("fileWords error reading %s: %v\n", path, err)
		return
	}
	//var words []string
	var startIdx int
	var hasLetter bool
	for i, b := range string(raw) {
		//non-word char case (end word):
		if !unicode.IsLetter(b) && !unicode.IsNumber(b) {
			//skip words with 0 letters
			if hasLetter {
				//push wloc entry
				s := string(raw)[startIdx:i]
				ent := wloc{key: hash([]byte(s)), fidx: len(dx.files)}
				dx.data = append(dx.data, ent)
				//reset word detection
				hasLetter = false
			}
			startIdx = -1
			continue
		}
		if unicode.IsLetter(b) {
			hasLetter = true
		}
		//start word, if not yet started
		if startIdx == -1 {
			startIdx = i
		}
	}
	dx.files = append(dx.files, path)
}
