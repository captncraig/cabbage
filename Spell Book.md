---
scratch: cabbage_scratch
---
```dataviewjs
let self = dv.current()
let scratch = dv.page(self["scratch"])
dv.header(2, "Spellbook for " + scratch["class"] + " / " + scratch["subclass"])
for (let group of dv.pages("#spell").groupBy(p => p.level)) {
    dv.header(3, group.key);
    dv.table(["Name", "Level", "Prepared", "Source"],
        group.rows
            .where(x => x[scratch.class] || x[scratch.subclass])
            .sort(x => x.title)
            .map(x => [
                x.file.link, 
                x.level, 
                x[scratch.subclass] ? "✅":
                "`INPUT[toggle:"+self.scratch+"#prepared_"+x.id+"]`",
                x.source
            ]))
}
```
