---
scratch: cabbage_scratch
---
```stats
items:
  - label: Spell Attack
    value: '+5'
  - label: Spell DC
    value: 20

grid:
  columns: 2
```

---

```consumable
items:
  - label: "Level 1"
    state_key: cab_spell_1
    uses: 4
  - label: "Level 2"
    state_key: cab_spell_2
    uses: 3
  - label: "Level 3"
    state_key: cab_spell_3
    uses: 3
  - label: "Level 4"
    state_key: cab_spell_4
    uses: 3
  - label: "Level 5"
    state_key: cab_spell_5
    uses: 3
  - label: "Level 6"
    state_key: cab_spell_6
    uses: 1
  - label: "Level 7"
    state_key: cab_spell_7
    uses: 1
  - label: "Level 8"
    state_key: cab_spell_8
    uses: 1
  - label: "Level 9"
    state_key: cab_spell_9
    uses: 1
```

```dataviewjs
let self = dv.current()
let scratch = dv.page(self["scratch"])
let prepared = dv.pages("#spell").where(x => x.level > 0 && !x[scratch.subclass] && scratch["prepared_"+x.id]);
dv.header(2, "Prepared spells (" + prepared.length +" / "+ scratch["max_prep"] +")")
let spells = dv.pages("#spell").where(x => x[scratch.subclass] || scratch["prepared_"+x.id])

dv.table(["Name", "Level", "Prepared"],
        spells
            .sort(x => x.title)
            .sort(x => x.level)
            
            .map(x => [
                x.file.link, 
                x.level, 
                x[scratch.subclass] ? "🍄":
                "`INPUT[toggle:"+self.scratch+"#prepared_"+x.id+"]`"
            ]))

```