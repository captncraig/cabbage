package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

const spacer = "\n>"

type Entry struct {
	Kind string `json:"type"`
}

func walkEntries(entries []json.RawMessage) (out []string) {
	for _, e := range entries {
		var s string
		if err := json.Unmarshal(e, &s); s != "" && err == nil {
			out = append(out, s+spacer)
			continue
		}
		var en Entry
		unmarshall(e, &en)
		out = append(out, walkEntry(e, en.Kind)...)
		continue
	}
	return out
}

type EntryListBase struct {
	Style string `json:"style"`
}

type EntryList struct {
	Style string          `json:"style"`
	Items []EntryListItem `json:"items"`
}

type EntryListItem struct {
	Kind    string   `json:"type"`
	Name    string   `json:"name"`
	Entries []string `json:"entries"`
}

type EntryListLite struct {
	Items []string `json:"items"`
}

type EntryEntries struct {
	Name    string            `json:"name"`
	Entries []json.RawMessage `json:"entries"`
}

func unmarshall(e json.RawMessage, v any) {
	if err := json.Unmarshal(e, v); err != nil {
		log.Fatalf("failed to unmarshal entry: %s", err)
	}
}

func walkEntry(e json.RawMessage, kind string) (out []string) {
	switch kind {
	case "list":
		elb := &EntryListBase{}
		unmarshall(e, elb)
		if elb.Style != "" {
			el := &EntryList{}
			unmarshall(e, el)
			for _, item := range el.Items {
				out = append(out, fmt.Sprintf("**%s:** %s"+spacer, item.Name, strings.Join(item.Entries, spacer)))
			}
		} else {
			ell := &EntryListLite{}
			unmarshall(e, ell)
			for _, item := range ell.Items {
				out = append(out, "-  "+item+spacer)
			}
		}
	case "entries":
		ee := &EntryEntries{}
		unmarshall(e, ee)
		out = append(out, fmt.Sprintf("**%s:**"+spacer, ee.Name))
		out = append(out, walkEntries(ee.Entries)...)
	case "table":
		et := &EntryTable{}
		unmarshall(e, et)
		out = append(out, walkTable(et)...)
		out = append(out, "")
	case "inset":
		// todo: what is this?

	default:
		log.Fatalf("unknown entry kind: %s %s", kind, string(e))
	}
	return
}

type EntryTable struct {
	Caption   string   `json:"caption"`
	ColStyles []string `json:"colStyles"`
	ColLabels []string `json:"colLabels"`
	Rows      [][]any  `json:"rows"`
}

func walkTable(et *EntryTable) []string {
	fmt.Println("!!!!")
	out := []string{}
	if et.Caption != "" {
		out = append(out, fmt.Sprintf("**%s:**"+spacer, et.Caption))
	}
	line := "|"
	for range et.ColLabels {
		line += "---|"
	}
	out = append(out, fmt.Sprintf("| %s |", strings.Join(et.ColLabels, " | ")))
	out = append(out, line)
	for _, row := range et.Rows {
		cells := []string{}
		for _, cell := range row {
			if rstr, ok := cell.(string); ok {
				cells = append(cells, rstr)
			} else if rarr, ok := cell.([]any); ok {
				cells = append(cells, rarr[1].(string))
			}
		}
		out = append(out, fmt.Sprintf("| %s |", strings.Join(cells, " | ")))
	}
	return out
}
