package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var TARGET_FILETYPES = []string{"json", "html", "csv", "py", "text", "c", "lua", "cpp", "sh", "txt", "xml", "yaml"}

var allFiletypes map[string]bool

func main() {
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("enter search term")
		s.Scan()
		fmt.Println("searching...")
		matches := search(s.Text())
		for _, m := range matches {
			fmt.Println(m)
		}
	}
}

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

var hashtable map[uint32]string

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
				//s := sraw[startIdx:i]
				if HASHTABLE {
					if hashtable == nil {
						hashtable = make(map[uint32]string)
					}
					hashtable[hash([]byte(sraw[startIdx:i]))] = sraw[startIdx:i]
				}
				ent := wloc{key: hash([]byte(sraw[startIdx:i])), fidx: len(dx.files)} //avoid copying string
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
	ndat := []wloc{dat[0]} //confusing
	for i := range dat[1:] {
		if dat[i] == dat[i+1] {
			continue
		}
		ndat = append(ndat, dat[i+1])
	}
	dx.data = ndat
	fmt.Printf("final size: %v wlocs\n", len(dx.data))
}

const DICTIONARY_MAX_SIZE = 5000000 //word count which triggers dumping to disk.
const DICTIONARY_MAX_COUNT = 30     //maximum number of dicts written to disk before indexer gives up.
const HASHTABLE = true              //save list of all words and hashes in a file, for debugging

// index directory. dumps data to index files when wloc count exceeds DICTIONARY_MAX_SIZE, creates a maximum of DICTIONARY_MAX_COUNT
// index files.  any dictionary data not dumped is returned.
// because current file indexing finishes befor dictionary is witten to disk, wloc count may exceed DICTIONARY_MAX_SIZE
func indexDir(root string) Dictionary {
	//print progress updates
	dx := Dictionary{}
	if HASHTABLE {
		hashtable = make(map[uint32]string)
	}
	var dcount int
	var stop bool
	var fullstop bool
	starter := make(chan bool) //used to coordinate writing dx to disk then restart of indexing process
	go func() {
		for {
			fmt.Printf("read %v files, stored %v wlocs\n", len(dx.files), len(dx.data))
			//write to disk when dx size limit is hit
			if len(dx.data) > DICTIONARY_MAX_SIZE {
				if dcount == DICTIONARY_MAX_COUNT {
					fmt.Println("hit DCOUNT_MAX")
					fullstop = true
					return
				}
				fmt.Printf("Dictionary hit size limit %d\n", DICTIONARY_MAX_SIZE)
				stop = true //flag to request indexer stop
				<-starter   //signal from indexer that dict is good to write
				sortDictionary(&dx)
				dedupDictionary(&dx)
				//f, _ := os.Create(fmt.Sprintf("dx%v.txt", dcount))
				fd, _ := os.Create(fmt.Sprintf("dx%v_data.data", dcount))
				fn, _ := os.Create(fmt.Sprintf("dx%v_fnames.txt", dcount))
				if HASHTABLE {
					hf, _ := os.Create(fmt.Sprintf("hashes%v.txt", dcount))
					for h, s := range hashtable {
						fmt.Fprintf(hf, "%.8x %s\n", h, s)
					}
					hashtable = make(map[uint32]string)
				}
				//dx.fPrint(f)
				dx.fWriteData(fd)
				dx.fWriteFilenames(fn)
				dcount++
				dx = Dictionary{}
				stop = false
				starter <- true //signal to indexer - continue indexing
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
			starter <- true //tell parent that indexing has stopped, proceed with write
			<-starter       //wait for parent to confirm continue OK
		}
		nm := f.Name()
		full := fmt.Sprintf("%v/%v", dir, nm)
		//don't index indexfiles *** TODO: make this more portable ***
		if strings.HasPrefix(full, "/Users/maxgara/Desktop/fs-search") {
			continue
		}
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

// get filenames containing s
func search(s string) []string {
	key := hash([]byte(s))
	var matches []string
	files, _ := os.ReadDir("/Users/maxgara/Desktop/fs-search")
	for _, f := range files {
		//if !strings.HasPrefix(f.Name(), "dx") {
		//continue
		//}
		if !strings.HasSuffix(f.Name(), "_fnames.txt") {
			continue
		}
		istr := strings.TrimSuffix(f.Name(), "_fnames.txt")
		istr = strings.TrimPrefix(istr, "dx")
		fmt.Printf("reading dict #%v \n", istr)
		dx := loadDictionary2(f.Name(), "dx"+istr+"_data.data")
		for _, d := range dx.data {
			if d.key == key {
				matches = append(matches, dx.files[d.fidx])
			}
		}
	}
	return matches
}

func loadDictionary(f string) Dictionary {
	data, err := os.ReadFile(f)
	if err != nil {
		fmt.Printf("couldn't read Dictionary from file %v: %v\n", f, err)
	}
	dataSplit := strings.Split(string(data), "\n  ***\n")
	pstrs := strings.Split(dataSplit[0], "\n") //path strings
	wstrs := strings.Split(dataSplit[1], "\n") //wloc strings
	var files []string
	for i := range pstrs {
		//ex pstr: 30: /Users/maxgara/Desktop/go-code/gobook/workspace/graph/rewrite/example-goal.html
		path := strings.Split(pstrs[i], ": ")[1]
		files = append(files, path)
	}
	var wls []wloc
	for i := range wstrs[:len(wstrs)-1] {
		splt := strings.Fields(wstrs[i])
		h, _ := strconv.ParseUint(splt[0], 16, 32)
		fidx, _ := strconv.ParseInt(splt[1], 10, 32)
		wl := wloc{key: uint32(h), fidx: int(fidx)}
		wls = append(wls, wl)
	}
	return Dictionary{files: files, data: wls}
}

func loadDictionary2(namefile, datafile string) Dictionary {
	var dx Dictionary
	nb, err := os.ReadFile(namefile)
	if err != nil {
		fmt.Printf("couldn't read Dictionary from file %v: %v\n", namefile, err)
	}
	db, err := os.ReadFile(datafile)
	if err != nil {
		fmt.Printf("couldn't read Dictionary from file %v: %v\n", datafile, err)
	}
	names := strings.Split(string(nb), "\n")
	dx.files = names[:len(names)-1] //drop the empty string after final newline
	//each line of data in db is 10 bytes long - see fWriteData
	for i := 0; i+10 < len(db); i += 10 {
		var key, fidx uint32
		_, err := binary.Decode(db[i:i+4], binary.NativeEndian, &key)
		if err != nil {
			fmt.Printf("error decoding data from %v: %v\n", datafile, err)
			return dx
		}
		_, err = binary.Decode(db[i+5:i+9], binary.NativeEndian, &fidx)
		if err != nil {
			fmt.Printf("error decoding data from %v: %v\n", datafile, err)
			return dx
		}
		dx.data = append(dx.data, wloc{key: key, fidx: int(fidx)})
	}
	return dx
}

func (dx *Dictionary) fWriteData(f io.Writer) {
	//buf: hash(4 byte) + space(1 byte) + fileidx(4 byte) + \n(1 byte)
	buf := make([]byte, 10)
	for _, d := range dx.data {
		_, err := binary.Encode(buf, binary.NativeEndian, struct {
			hash    uint32
			space   byte
			fidx    uint32
			newline byte
		}{d.key, ' ', uint32(d.fidx), '\n'})
		if err != nil {
			fmt.Printf("fWriteData err: %v\n", err)
			return
		}
		f.Write(buf)
	}
}

func (dx *Dictionary) fWriteFilenames(f io.Writer) {
	for _, name := range dx.files {
		fmt.Fprintf(f, "%s\n", name)
	}
}

// completion node
type compNode struct {
	c       rune //last char in partial word
	next    int  //index of child compnode array in dict
	nextlen int  //len of child compnode array
}

func WildCardDict(fs []string) {
	wcd := []compNode{}
	set := "abcdefghijklmnopqrstuvwxyz" //replace with all unicode code points in fs
	for _, c := range set {
		wcd = append(wcd, compNode{c: c})
	}
}
