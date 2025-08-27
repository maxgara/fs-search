package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"time"
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
	sraw := string(raw)
	var startIdx int
	var hasLetter bool
	for i, b := range sraw {
		//non-word char case (end word):
		if !unicode.IsLetter(b) && !unicode.IsNumber(b) {
			//skip words with 0 letters
			if hasLetter {
				//push wloc entry
				s := sraw[startIdx:i]
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

// implement sort.Interface
func (dx *Dictionary) Less(i, j int) bool {
	return dx.data[i].key < dx.data[j].key || ((dx.data[i].key == dx.data[j].key) && (dx.data[i].fidx < dx.data[j].fidx))
}
func (dx *Dictionary) Swap(i, j int) {
	dx.data[i], dx.data[j] = dx.data[j], dx.data[i]
}
func (dx *Dictionary) Len() int {
	return len(dx.data)
}

// convenience func
func (dx Dictionary) String() string {
	var s string
	for i, name := range dx.files {
		s += fmt.Sprintf("%d: %s\n", i, name)
	}
	s += "  ***\n"
	for _, d := range dx.data {
		s += fmt.Sprintf("%.8x %v\n", d.key, d.fidx)
	}
	return s
}
func (dx Dictionary) fPrint(f io.Writer) {
	for i, name := range dx.files {
		io.WriteString(f, fmt.Sprintf("%d: %s\n", i, name))
	}
	io.WriteString(f, "  ***\n")
	for _, d := range dx.data {
		io.WriteString(f, fmt.Sprintf("%.8x %v\n", d.key, d.fidx))
	}
}

// temporary implementation
func sortDictionary(dx *Dictionary) {
	fmt.Println("sorting dictionary")
	sort.Sort(dx)
	fmt.Println("sorted")
}

// temporary implementation?
func dedupDictionary(dx *Dictionary) {
	fmt.Println("deduping")
	fmt.Printf("initial size: %v wlocs\n", len(dx.data))
	dat := dx.data
	ndat := []wloc{dat[0]}
	for i := range dat[1:] {
		if dat[i] == dat[i+1] {
			continue
		}
		ndat = append(ndat, dat[i+1])
	}
	dx.data = ndat
	fmt.Printf("final size: %v wlocs\n", len(dx.data))
}

const DICTIONARY_MAX_SIZE = 1000000
const DCOUNT_MAX = 3

func dictFromDir(root string) Dictionary {
	//print progress updates
	dx := Dictionary{}
	var dcount int
	var stop bool
	var fullstop bool
	starter := make(chan bool) //used to coordinate writing dx to disk and restart of indexing process
	go func() {
		for {
			fmt.Printf("read %v files, stored %v wlocs\n", len(dx.files), len(dx.data))
			if len(dx.data) > DICTIONARY_MAX_SIZE {
				if dcount == DCOUNT_MAX {
					fmt.Println("hit DCOUNT_MAX")
					fullstop = true
					return
				}
				fmt.Printf("Dictionary hit size limit %d\n", DICTIONARY_MAX_SIZE)
				stop = true
				<-starter //wait for indexing to stop
				sortDictionary(&dx)
				dedupDictionary(&dx)
				f, _ := os.Create(fmt.Sprintf("dx%v.txt", dcount))
				dx.fPrint(f)
				dcount++
				dx = Dictionary{}
				stop = false
				starter <- true //continue indexing
			}
			time.Sleep(time.Second / 10)
		}
	}()
	rdfd(&dx, root, &stop, starter, &fullstop)
	sortDictionary(&dx)
	dedupDictionary(&dx)
	return dx
}
func rdfd(dx *Dictionary, dir string, stop *bool, starter chan bool, fullstop *bool) {
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("ReadWalk dir: %s: %v\n", dir, err)
	}
	for _, f := range files {
		if *fullstop {
			return
		}
		if *stop {
			starter <- true //tell parent that indexing has stopped
			<-starter       //wait for parent to confirm continue OK
		}
		nm := f.Name()
		full := fmt.Sprintf("%v/%v", dir, nm)
		if f.IsDir() {
			rdfd(dx, full, stop, starter, fullstop)
			continue
		}
		for i := len(nm) - 2; i >= 0; i-- {
			if nm[i] != '.' {
				continue
			}
			ft := nm[i+1:] // filetype eg: "csv"
			if slices.Contains(TARGET_FILETYPES, ft) {
				addFileKeys(full, dx)
			}
			break
		}
	}
}
