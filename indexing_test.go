package main

import (
	"fmt"
	"os"
	"testing"
)

func TestReadWalk(t *testing.T) {
	files := ReadWalk("/Users/maxgara/Desktop")
	//	for k, _ := range allFiletypes {
	//		fmt.Println(k)
	//	}
	fmt.Printf("targets: %d\n", len(files))
}

func TestFileWords(t *testing.T) {
	words := fileWords("/Users/maxgara/Desktop/test.c")
	fmt.Println(words)
}

func TestHash(t *testing.T) {
	words := fileWords("/Users/maxgara/Desktop/test.c")
	for _, w := range words {
		bts := []byte(w)
		h := hash(bts)
		fmt.Printf("%s => %x\n", w, h)
	}
}

func TestAddFileKeys(t *testing.T) {
	var dx Dictionary
	addFileKeys("/Users/maxgara/Desktop/test.c", &dx)
	addFileKeys("/Users/maxgara/Desktop/rules.txt", &dx)
	fmt.Println(dx)
}

func TestSortDictionary(t *testing.T) {
	var dx Dictionary
	addFileKeys("/Users/maxgara/Desktop/test.c", &dx)
	addFileKeys("/Users/maxgara/Desktop/rules.txt", &dx)
	sortDictionary(&dx)
	fmt.Println(dx)
}

func TestDedupDictionary(t *testing.T) {
	var dx Dictionary
	addFileKeys("/Users/maxgara/Desktop/test.c", &dx)
	addFileKeys("/Users/maxgara/Desktop/rules.txt", &dx)
	sortDictionary(&dx)
	dedupDictionary(&dx)
	fmt.Println(dx)
}
func TestDictFromDir(t *testing.T) {
	dx := indexDir("/Users/maxgara/Desktop")
	f, _ := os.Create("dedupDictionaryOutput.txt")
	dx.fPrint(f)
	fmt.Println("TEST DONE")
	//dx.fPrint(os.Stdout)
	//fmt.Println(dx)
}

// use binary filetype
func TestDictFromDir2(t *testing.T) {
	dx := indexDir("/Users/maxgara")
	fd, _ := os.Create("dedupDictionaryOutput.txt")
	fn, _ := os.Create("dedupDictionaryOutput.txt")
	dx.fWriteData(fd)
	dx.fWriteFilenames(fn)
	fmt.Println("TEST DONE")
	//dx.fPrint(os.Stdout)
	//fmt.Println(dx)
}

func TestLoadDictionary(t *testing.T) {
	dx := loadDictionary("dx6.txt")
	dx.fPrint(os.Stdout)
}

func TestSearch(t *testing.T) {
	matches := search("thisisatestfileforfssearch")
	for _, s := range matches {
		fmt.Println(s)
	}
}

func TestFWriteData(t *testing.T) {
	var dx Dictionary
	addFileKeys("/Users/maxgara/Desktop/test.c", &dx)
	fmt.Println(dx)
	f, _ := os.Create("fwritedatatest.txt")
	dx.fWriteData(f)
}

func TestFWriteFilenames(t *testing.T) {
	var dx Dictionary
	addFileKeys("/Users/maxgara/Desktop/test.c", &dx)
	fmt.Println(dx)
	f, _ := os.Create("fwritefilenamestest.txt")
	dx.fWriteFilenames(f)
}

// depends on the two prior tests lol
func TestLoadDictionary2(t *testing.T) {
	dx := loadDictionary2("fwritefilenamestest.txt", "fwritedatatest.txt")
	fmt.Println(dx)
}

// depends on the two prior tests lol
func TestWildCardDict(t *testing.T) {
	dx := loadDictionary2("fwritefilenamestest.txt", "fwritedatatest.txt")
 WildCardDict(dx.files)
}

// completion node
type compNode struct {
	c       rune //last char in partial word
	next    int  //index of child compnode array in dict
	nextlen int  //len of child compnode array
}

type compTable []compNode

// create a wildcard dictionary to find partial matches for string fragments
func WildCardDict(fs []string) {
	wcd := []compNode{}
	set := "abcdefghijklmnopqrstuvwxyz" //replace with all unicode code points in fs
	for _, c := range set {
		wcd = append(wcd, compNode{c: c})
	}
}

// add next and nextlen attributes for node, using partial word strpart and considering word possibilities fs
func rwcd(fs []string, t *compTable, strpart string, node *compNode) {
	//group words in fs by letter occurring after strpart
	//in some cases there will be more than one letter that meets this criteria - in this case, place word in both groups
	part := []rune(strpart)
	g := make(map[rune][]string)
	for _, str := range fs {
		s := []rune(str)
		i := 0
		j := 0
		for ; j < len(s); j++ {
			//if part does not match, start match over further in f
			if part[i] != s[j] {
				j -= i
				i = 0
				continue
			}
			i++
			//partial match case
			if i < len(part) {
				continue
			}
			//if part has fully matched, place f in map
			//special case where part is at the end of f (terminal match):
			if j+1 >= len(s) {
				g[0] = append(g[0], str) // unicode null
				break
			}
			//non-terminal match
			nextc := s[j+1]
			g[nextc] = append(g[nextc], str)
			i = 0
		}
	}
	//first fill in t with nextlen semi-blank compnodes, then recursively call rwcd on these nodes to add attributes
	nodes := []compNode{}
	nextfs := [][]string{}
	for c, strs := range g {
		cn := compNode{c: c}
		nodes = append(nodes, cn)
		nextfs = append(nextfs, strs)
	}
	*t = append(*t, nodes...)
	for i := range nodes {
		c := &nodes[i]
		rwcd(nextfs[i], t, strpart+string(c.c), c)
	}
	//assign attrs. to node
	node.next = len(*t)
	node.nextlen = len(g)
}

