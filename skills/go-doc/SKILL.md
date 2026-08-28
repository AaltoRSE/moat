---
name: go-doc
description: Guidelines and best practices for writing and reviewing Go doc comments.
---

# Go Doc Comment Skill

Write and review Go doc comments.

---

## Core rules

- Every exported name **must** have a doc comment.
- A doc comment sits **immediately before** the declaration with no blank line between them.
- Use full sentences. Start with the name of the thing being documented.
- Do not document internal implementation details in doc comments; put those inside the function body.

---

## Comment form by declaration kind

### Package

```go
// Package path implements utility routines for manipulating slash-separated paths.
//
// The path package should only be used for paths separated by forward slashes,
// such as the paths in URLs. This package does not deal with Windows paths;
// to manipulate operating system paths, use the [path/filepath] package.
package path
```

- First sentence begins with `"Package <name>"`.
- Multi-file packages: put the package comment in exactly one file.

### Command (main package)

```go
/*
Gofmt formats Go programs.
It uses tabs for indentation and blanks for alignment.

Usage:

    gofmt [flags] [path ...]
*/
package main
```

- First sentence starts with the program name (capitalized, as it begins the sentence).
- Include a `Usage:` section with an indented command line when the command has flags.

### Type

```go
// A Reader serves content from a ZIP archive.
type Reader struct { ... }
```

- Describe what an instance **represents or provides**.
- Document zero-value behavior when it is meaningful.
- Document concurrency safety when stronger than the default (single goroutine).
- Explain exported fields either in the type comment or with per-field comments.

### Function / Method

```go
// Quote returns a double-quoted Go string literal representing s.
func Quote(s string) string { ... }

// HasPrefix reports whether s begins with prefix.
func HasPrefix(s, prefix string) bool { ... }
```

- Focus on **what is returned** (or what is done, for side-effect functions).
- Use `"reports whether"` for boolean-returning predicates; omit `"or not"`.
- Refer to parameters and results by name directly — no backtick quoting needed.
- Document special cases (errors, panics, edge inputs) explicitly.
- Do not describe the algorithm; describe the contract.

### Const / Var

```go
// Version is the Unicode edition from which the tables are derived.
const Version = "13.0.0"

// Generic file system errors tested with errors.Is.
var (
    ErrInvalid    = errInvalid()    // "invalid argument"
    ErrPermission = errPermission() // "permission denied"
)
```

- A group of related consts/vars may share one leading doc comment.
- Individual items in a group use short end-of-line comments.
- Typed constants grouped under a type usually rely on the type's doc comment.

---

## Syntax reference

### Headings

```
// # Numeric Conversions
```

A `# ` prefix on an otherwise blank-line-separated, single-line paragraph.

### Links

```
// See [RFC 7159] for details.
//
// [RFC 7159]: https://tools.ietf.org/html/rfc7159
```

Put URL definitions at the end of the comment. Plain URLs are auto-linked.

### Doc links

```
// ReadFrom appends to [Buffer] and returns [io.EOF] on completion.
```

- Same-package symbol: `[Name]` or `[Name.Method]`
- Other package: `[pkg.Name]` or full import path when ambiguous

### Lists

```
// Clean applies the following rules:
//
//  1. Replace multiple slashes with a single slash.
//  2. Eliminate each . path name element.
//
// It returns "." for an empty result.
```

Bullet list: `  - item` (two-space indent, dash marker).
Numbered list: `  1.` or `  1)` (one-space indent).
No nested lists — flatten or mix bullet/number markers as a workaround.

### Code blocks

Indent by one tab (or use consistent indentation that gofmt will normalise):

```
// Example:
//
//  result := compute(x, y)
```

### Notes / TODOs

```
// TODO(username): refactor to use standard library context
// BUG(username): not cleaned up on error path
```

### Deprecations

```
// Deprecated: RC4 is cryptographically broken. Use [crypto/aes] instead.
```

The paragraph starting with `Deprecated: ` hides the symbol on pkg.go.dev.

---

## Common mistakes to avoid

| Mistake | Fix |
|---|---|
| Indenting a wrapped paragraph line | Unindent continuation lines |
| Not indenting list continuation lines | Indent continuation lines 4 spaces |
| Using `//` without a space before text | Always write `// Text` |
| Starting a func comment with `This function…` | Start with the symbol name: `// Foo returns…` |
| Documenting the algorithm instead of the contract | Move algorithm notes inside the function body |
| Using nested lists | Flatten, or mix bullet/numbered markers |
| Placing a blank line between a doc comment and its declaration | Remove the blank line |

---

## Checklist when writing or reviewing

- [ ] Every exported symbol has a doc comment.
- [ ] Comment starts with the symbol name.
- [ ] Full sentences with proper punctuation.
- [ ] Boolean predicates use "reports whether".
- [ ] Special cases (panics, errors, edge inputs) are documented.
- [ ] Cross-references use `[Symbol]` or `[pkg.Symbol]` doc links.
- [ ] No implementation details in the public doc comment.
- [ ] Deprecation notice uses `Deprecated: ` prefix.
- [ ] Code blocks and lists are correctly indented.
