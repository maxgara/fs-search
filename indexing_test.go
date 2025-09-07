package main

import (
	"fmt"
	"os"
	"testing"
	"time"
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

func TestWildCardDict(t *testing.T) {
	dx := loadDictionary2("dx1_fnames.txt", "fwritedatatest.txt")
	fmt.Println(dx.files)
	<-time.After(5 * time.Second)
	wd := WildCardDict(dx.files)
	fmt.Println(wd.table)
}
