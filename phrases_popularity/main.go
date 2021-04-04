package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
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
	chanPopulated := make(chan string) // to push paragraphs into
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
        var prgrphOccrncsUniqueAggr []occurence
	for v := range chanProcessed{
		skip := false
                for i, u := range prgrphOccrncsUniqueAggr {
                        if v.Phrase == u.Phrase {
                                prgrphOccrncsUniqueAggr[i].Count += v.Count // sum up Counts
                                skip = true
                        }
                }
                if !skip {
                        prgrphOccrncsUniqueAggr = append(prgrphOccrncsUniqueAggr, v)
                }
	}

	sort.Sort(ByCount(prgrphOccrncsUniqueAggr))

	// dump first 100 structs sorted in descending way into json
	outputLen := len(prgrphOccrncsUniqueAggr)
	if outputLen > 100 {outputLen = 100}

	file, _ := json.MarshalIndent(prgrphOccrncsUniqueAggr[:outputLen], "", " ")
	_ = ioutil.WriteFile("output.json", file, 0644)

	fmt.Printf("\n== Results stored in output.json ===\n")
}

// ==============
func errCheck(e error) {
	if e != nil {
		panic(e)
	}
}

// ==============

//  take substring, split by paragraph
func populate(d string, c chan<- string) {
	prgrphs := strings.Split(d, "\n\n")

	for i := range prgrphs {
		prgrph := prgrphs[i]
		if len(prgrph) == 0 {
			continue
		} // to skip empty elements
		c <- prgrph
	}
	close(c)
}

// to return channel of prgrphOccrncsUniqueMerged
func process(cIn <-chan string, cOut chan<- occurence) {
	var wg sync.WaitGroup
	for p := range cIn {
		if len(p) == 0 {
			continue
		} // to skip empty elements
		wg.Add(1)
		go func(prgrph string) {
			fmt.Println("Goroutines:", runtime.NumGoroutine())

			var prgrphOccrncs []occurence
			var prgrphOccrncsUnique []occurence

			// replcing characters and extra spaces
			reg, err := regexp.Compile(`[^\w+]`)
		        errCheck(err)
		        prgrph = reg.ReplaceAllString(prgrph, " ")
                        reg, err = regexp.Compile(`\s{2,}`)
                        errCheck(err)
                        prgrph = reg.ReplaceAllString(prgrph, " ")
			prgrph = strings.ToLower(prgrph)

			words := strings.Fields(prgrph) // to construct searching phrases

			for j := 2; j < len(words); j++ { // starting on 2 to catch 3-words phrase from beginning
				substr := occurence{
					Phrase: strings.Join([]string{words[j-2], words[j-1], words[j]}, " "),
					Count:  0,
				}
				lookupPhraseRegexp := regexp.MustCompile(substr.Phrase)
				phraseMatches := lookupPhraseRegexp.FindAllStringIndex(prgrph, -1)
				substr.Count = len(phraseMatches)
				prgrphOccrncs = append(prgrphOccrncs, substr)
			}

			// now we might have duplicate structs after processing paragraph
			for _, v := range prgrphOccrncs {
				skip := false
				for _, u := range prgrphOccrncsUnique {
					if v == u {
						skip = true
						break
					}
				}
				if !skip {
					prgrphOccrncsUnique = append(prgrphOccrncsUnique, v)
				}
			}

			// push to channle, fan-in
			for _, v := range prgrphOccrncsUnique {
				cOut <- v
			}

			fmt.Printf("paragraph complete\n")
			wg.Done()
		}(p)
	}

	wg.Wait()
	close(cOut)
}
