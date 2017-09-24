# Argonaut [![GoDoc](https://godoc.org/github.com/ghetzel/argonaut?status.svg)](https://godoc.org/github.com/ghetzel/argonaut)
A library for generating and parsing complex command line invocations in a structured manner.

## Overview

Argonaut is a library that makes working with complex command line utilities simpler.  Instead of using complex application logic and string slices for building up command line invocations, you can model shell commands using Golang structs, then using Argonaut, marshal those structs into valid strings that can be used to shell out.

## Example

Let's look at how you would model the standard coreutils `ls` command.  Here is a minimal struct that defines some of the options we want to work with:

```golang
type ls struct {
    Command       argonaut.CommandName `argonaut:"ls"`
    All           bool                 `argonaut:"all|a"`
    BlockSize     int                  `argonaut:"block-size,long"`
    LongFormat    bool                 `argonaut:"l"`
    HumanReadable bool                 `argonaut:"human-readable|h"`
    Paths         []string             `argonaut:",positional"`
}
```

And here are some things you can do using _Argonaut_:

```golang
// build our command invocation using native types
ls_alh := &ls{
    All:           true,
    LongFormat:    true,
    HumanReadable: true,
    Paths: []string{
        `/foo`,
        `/bar/*.txt`,
        `/baz/`,
    },
})

// Marshal into a []byte slice
command, _ := argonaut.Marshal(ls_alh)

fmt.Println(string(command))
// Output: "ls --all -l --human-readable /foo /bar/*.txt /baz/"

// Parse into a []string slice
fmt.Println(strings.Join(argonaut.MustParse(ls_alh), `|`))
// Output: "ls|--all|-l|--human-readable|/foo|/bar/*.txt|/baz/"


// Runs the command, returns the standard output
stdout, err := argonaut.MustCommand(ls_alh).Output()
// Returns: the output of the command, <nil> (or non-nil if the command failed)
```

## Using the `argonaut` Struct Tag

The `argonaut:` tag can be used to tell Argonaut how fields in a struct should be converted into command line arguments.  Except for positional arguments, the first part of the tag value (everything before the first comma) specifies the parameter name as it will appear in the generated command line.  Both long (e.g.: `--argument`) and short (e.g.: `-a`) labels are supported.  If both variants are valid, they should be separated by a pipe (`|`), with the default form occurring first (e.g.: `argument|a`).

Everything after the first comma represents additional configuration used to fine-tune the presentation of the parameter.

| Configuration Item | Description               |
| ------------------ | ------------------------- |
| `long`             | The parameter only supports a long-form argument. |
| `short`            | The parameter only supports a short-form argument. |
| `positional`       | The field represents a positional argument.  Can be a slice type. |
| `required`         | The parameter must be specified (cannot contain a zero value). |
| `suffixprev`       | The value of the field is not a standalone parameter, but is instead a modifier for the parameter immediately preceding the field.  The value will be concatenated with the previous parameter name, joined using the value of the `delimiters` configuration item.  The `delimiter` defaults to a single space (" "). |
| `delimiters=[...]` | Specifies a sequence of characters that should be used to join parameter name modifiers (specified by `suffixprev`).  See below for an example. |


### Example Usage for `suffixprev` and `delimiters`

```
type ComplexThing struct {
    Command    argonaut.CommandName `argonaut:"mycmd"`
    Filter     string               `argonaut:"filter"`
    FilterType string               `argonaut:",suffixprev,delimiters=[:]`
}

argonaut.MustCommand(ComplexThing{
    Filter:     `testing`,
    FilterType: `audio`,
})

// Returns: "mycmd --filter:audio testing"
```

## Rationale

This approach is useful in sitations where you are working with incredibly complex commands whose argument structures are very dynamic and nuanced.  Some examples that come to mind are [`ffmpeg`](https://ffmpeg.org/ffmpeg.html), [`vlc`](https://wiki.videolan.org/VLC-1-1-x_command-line_help/), and [`uwsgi`](https://uwsgi-docs.readthedocs.io/en/latest/).

In each circumstance, a fully-featured library exists that can be natively integrated with, but the sheer number of options implemented in the CLI tools and the logic therein that you would need to replicate makes shelling out a very attractive option.  But building the command line is often itself a non-trivial task.  Argonaut was built to make this process easier.
