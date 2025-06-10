---
scratch: cabbage_scratch
---
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
                x[scratch.subclass] ? "✅":
                "`INPUT[toggle:"+self.scratch+"#prepared_"+x.id+"]`"
            ]))

```