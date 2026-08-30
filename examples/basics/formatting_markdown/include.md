# Include Directive

Place `{include:path/to/fragment.md}` in a `_template.md` to embed a shared file at that position. The path is relative to the template's own directory.

Included files are expanded recursively, so an included file may itself use further `{include:}` directives and `{key}` placeholders.

See the *Input* category for a working example.

---

Template `_template.md`:
````markdown
# {name}

{description}

{include:_footer.md}
````

Fragment `_footer.md`:
````markdown
---
*Shared content appearing on every card.*
````
