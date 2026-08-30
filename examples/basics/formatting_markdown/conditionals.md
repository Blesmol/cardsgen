# Conditional Directive

Use `{if:key}...{endif}` to show a block only when a key is present and non-empty. Add `{else}` for an alternative branch. Use `{if:key=value}` to check for a specific value.

See the *Input* category for a working example.

---

Presence check — show a block only when the key exists:
````markdown
{if:note}
> **Note:** {note}
{endif}
````

Value check with else — choose between two branches:
````markdown
{if:difficulty=hard}
**Difficulty:** Hard
{else}
**Difficulty:** Normal
{endif}
````
