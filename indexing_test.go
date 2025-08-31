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
	dx := indexDir("/Users/maxgara")
	f, _ := os.Create("dedupDictionaryOutput.txt")
	dx.fPrint(f)
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
