package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var includeSources = map[string]bool{
	"phb": true,
	"tce": true,
	"xge": true,
}

const dir5e = `../../5etools-src/data`

func main() {
	loadSources()
	idx, err := os.ReadFile(filepath.Join(dir5e, "spells/index.json"))
	if err != nil {
		log.Fatal(err)
	}
	srcs := map[string]string{}
	err = json.Unmarshal(idx, &srcs)
	if err != nil {
		log.Fatal(err)
	}

	for name, url := range srcs {
		if !includeSources[strings.ToLower(name)] {
			continue
		}
		os.MkdirAll(fmt.Sprintf("../Spells/%s", name), 0755)
		spells := SpellList{}
		file, err := os.ReadFile(filepath.Join(dir5e, "spells", url))
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(file, &spells)
		if err != nil {
			log.Fatal(err)
		}
		for _, spell := range spells.Spells {
			fmt.Println(spell.Name)
			err := os.WriteFile(fmt.Sprintf("../Spells/%s/%s.md", name, spell.filename()), []byte(spell.Markdown()), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

type SpellList struct {
	Spells []Spell `json:"spell"`
}
type Spell struct {
	Name               string            `json:"name"`
	Source             string            `json:"source"`
	Level              int               `json:"level"`
	School             string            `json:"school"`
	Entries            []json.RawMessage `json:"entries"`
	EntriesHigherLevel []json.RawMessage `json:"entriesHigherLevel"`
	Time               []SpellTime       `json:"time"`
	Duration           []SpellDuration   `json:"duration"`
	Components         SpellComponents   `json:"components"`
}
type SpellTime struct {
	Number    int    `json:"number"`
	Unit      string `json:"unit"`
	Condition string `json:"condition"`
}
type SpellDuration struct {
	Type     string `json:"type"`
	Duration struct {
		Type   string `json:"type"`
		Amount int    `json:"amount"`
		UpTo   bool   `json:"upTo"`
	} `json:"duration"`
	Concentration bool     `json:"concentration"`
	Ends          []string `json:"ends"`
}
type SpellComponents struct {
	Verbal   bool `json:"v"`
	Somatic  bool `json:"s"`
	Material any  `json:"m"`
}

func (sc SpellComponents) String() string {
	parts := []string{}
	if sc.Verbal {
		parts = append(parts, "V")
	}
	if sc.Somatic {
		parts = append(parts, "S")
	}
	if sc.Material != nil {
		if ps, ok := sc.Material.(string); ok {
			parts = append(parts, fmt.Sprintf("M(%s)", ps))
		} else {
			obj := sc.Material.(map[string]any)
			parts = append(parts, fmt.Sprintf("M(%s)", obj["text"].(string)))
		}
	}
	return strings.Join(parts, ", ")
}

func (s Spell) Markdown() string {
	buf := bytes.Buffer{}
	if err := spellTpl.Execute(&buf, s); err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func (s Spell) TimeString() string {
	cond := ""
	if s.Time[0].Condition != "" {
		cond = fmt.Sprintf(" %s", s.Time[0].Condition)
	}
	return fmt.Sprintf("%d %s%s", s.Time[0].Number, s.Time[0].Unit, cond)
}

func (s SpellDuration) DurationString() string {
	base := ""
	switch s.Type {
	case "instant":
		base = "instantaneous"
	case "timed":
		base = fmt.Sprintf("%d %s", s.Duration.Amount, s.Duration.Type)
		if s.Duration.UpTo {
			base = "(up to) " + base
		}
	case "permanent":
		base = "until dispelled"
	case "special":
		base = "special"
	default:
		log.Fatalf("unknown duration type: %s", s.Type)
	}
	if s.Concentration {
		base = fmt.Sprintf("%s (concentration)", base)
	}
	return base
}

var schoolLookup = map[string]string{
	"A": "Abjuration",
	"C": "Conjuration",
	"D": "Divination",
	"T": "Transmutaion",
	"E": "Enchantment",
	"V": "Evocation",
	"I": "Illusion",
	"N": "Necromancy",
}

func (s Spell) LevelLine() string {
	if schoolLookup[s.School] == "" {
		log.Fatal("unknown school: ", s.School)
	}
	if s.Level == 0 {
		return fmt.Sprintf("%s cantrip", schoolLookup[s.School])
	}
	return fmt.Sprintf("%s level %s", getOrdinalNumber(s.Level), schoolLookup[s.School])
}

func (s Spell) CleanEntries() []string {
	return append(cleanEntries(walkEntries(s.Entries)), cleanEntries(walkEntries(s.EntriesHigherLevel))...)
}

func (s Spell) Lists() []string {
	lists := []string{}
	for l := range AllSpellSources[strings.ToLower(s.Name)] {
		lists = append(lists, l)
	}
	sort.Strings(lists)
	return lists
}

func (s Spell) Alias() string {
	return alias("spell", s.Name)
}

func alias(kind, name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), " ", "_")
	return fmt.Sprintf("%s_%s", kind, name)
}

func cleanEntries(entries []string) []string {
	for i, s := range entries {
		entries[i] = cleanTxt(s)
	}
	return entries
}

func cleanTxt(txt string) string {
	var matches = extractRegex.FindAllStringSubmatch(txt, -1)
	for _, match := range matches {
		var replacement = ""
		switch kind := match[1]; kind {
		case "spell":
			replacement = fmt.Sprintf(`[[%s]]`, match[2])
		case "chance":
			replacement = fmt.Sprintf(`%s%%`, match[2])
		case "classFeature", "dc", "damage", "dice", "book", "skill", "status", "note", "filter", "item", "b", "hit", "quickref", "scaledamage", "scaledice":
			// simple replacements
			//{@dc 10}
			//{@dice 1d6}
			//{@damage 12d6}
			replacement = "**" + match[2] + "**"
		case "i":
			replacement = "_" + match[2] + "_"
		case "action", "variantrule", "condition", "hazard", "creature", "sense", "race":
			// possilbe reference links
			replacement = "**" + match[2] + "**"
		default:
			log.Fatal("unknown kind for tag replacement: ", kind)
		}
		if replacement != "" {
			txt = strings.ReplaceAll(txt, match[0], replacement)
		}
	}
	return txt
}

func IDString(s string) string {
	s = strings.ReplaceAll(strings.ToLower(s), " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "'", "_")
	return s
}
func (s Spell) IDString() string {
	return IDString(s.Name)
}
func (s Spell) filename() string {
	n := strings.ReplaceAll(s.Name, "/", "_")
	return n
}
func getOrdinalNumber(n int) string {
	if n >= 11 && n <= 13 {
		return fmt.Sprintf("%dth", n)
	}

	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

var spellTpl = template.Must(template.New("spell").Parse(`---
title: {{.Name}}
source: {{.Source}}
level: {{.Level}}
school: {{.School}}
id: {{.IDString}}
verbal: {{.Components.Verbal}}
somatic: {{.Components.Somatic}}
material: {{ne .Components.Material nil}}
aliases:
  - {{.IDString}}
tags:
  - spell
{{range .Lists}}{{.}}: true
{{end}}
---
>[!tip] {{.Name}}
>
> *{{.LevelLine}}*
> *Casting Time:* {{.TimeString}}
{{range .Duration}}> *Duration:* {{.DurationString}}
{{- end}}
> *Components:* {{.Components.String}}
>
{{range .CleanEntries}}>{{.}}
{{end}}
`))

var extractRegex = regexp.MustCompile(`(?mi)\{@([a-z]+)\s+([^\|\}]+)(\|[^\}]*)?\}`)
