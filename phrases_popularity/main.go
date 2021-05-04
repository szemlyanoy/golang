package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type occurence struct {
	Phrase string
	Count  int
}

var mu sync.Mutex
var prgrphOccrncsUniqueMerged []occurence

type ByCount []occurence

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Filename must be passed")
		return
	}
	chanPopulated := make(chan string)    // to push paragraphs into
	chanProcessed := make(chan occurence) // to push paragraphs into

	var buff bytes.Buffer
	var data string
	var err error
	for i := 1; i < len(os.Args); i++ {
		file, _ := ioutil.ReadFile(os.Args[i])
		buff.WriteString(string(file))
	}
	data = buff.String()
	errCheck(err)

	files, err := filepath.Glob("*.json")
	errCheck(err)
	for _, f := range files {
		err := os.Remove(f)
		errCheck(err)
	}

	// ##########################################
	go populate(data, chanPopulated)
	go process(chanPopulated, chanProcessed)

	// aggregation. Range over prgrphOccrncsUniqueMerged and find matching struct and aggregate counts
	prgrphOccrncsUniqueAggr := make(map[string]int)
	for v := range chanProcessed {
		prgrphOccrncsUniqueAggr[v.Phrase] += v.Count

	}
	// sorting by Count
	res := make(ByCount, len(prgrphOccrncsUniqueAggr))
	var i int
	for k, v := range prgrphOccrncsUniqueAggr {
		res[i] = occurence{k, v}
		i++
	}
	sort.Sort(ByCount(res)) // sorted

	// dump first 100 occurences sorted in descending way into json
	outputLen := len(res)
	if outputLen > 100 {
		outputLen = 100
	}
	resJson, _ := json.MarshalIndent(res[:outputLen], "", " ")
	_ = ioutil.WriteFile("output.json", resJson, 0644)

	fmt.Printf("\n== Results stored in output.json ===\n")
}
