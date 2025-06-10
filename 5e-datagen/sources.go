package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SourceList map[string]map[string]SpellSources

type SpellSources struct {
	Class        map[string]map[string]any                       `json:"class"`
	Subclass     map[string]map[string]map[string]map[string]any `json:"subclass"`
	ClassVariant map[string]map[string]any                       `json:"classVariant"`
}

var AllSpellSources = map[string]map[string]bool{}

func loadSources() {
	dat, err := os.ReadFile(filepath.Join(dir5e, "generated/gendata-spell-source-lookup.json"))
	if err != nil {
		log.Fatal(err)
	}
	top := SourceList{}
	unmarshall(dat, &top)
	allBooks := []string{}
	for book := range top {
		allBooks = append(allBooks, book)
	}
	sort.Strings(allBooks)

	for _, book := range allBooks {
		spells := top[book]
		if !includeSources[strings.ToLower(book)] {
			continue
		}
		for spell, sources := range spells {
			// last book wins, xphb I hope
			AllSpellSources[spell] = map[string]bool{}
			for src, data := range sources.Class {
				if !includeSources[strings.ToLower(src)] {
					continue
				}
				for class := range data {
					AllSpellSources[spell][IDString(class)] = true
				}
			}
			for src, data := range sources.ClassVariant {
				if !includeSources[strings.ToLower(src)] {
					continue
				}
				for class := range data {
					AllSpellSources[spell][IDString(class)] = true
				}
			}
			for src, data := range sources.Subclass {
				if !includeSources[strings.ToLower(src)] {
					continue
				}
				for class, srcs := range data {
					for src2, subs := range srcs {
						if !includeSources[strings.ToLower(src2)] {
							continue
						}
						for sub := range subs {
							AllSpellSources[spell][IDString(class)+"_"+IDString(sub)] = true
						}
					}
				}
			}
		}
	}
	fmt.Println(AllSpellSources)

}
