package main

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

func errCheck(e error) {
	if e != nil {
		panic(e)
	}
}

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
