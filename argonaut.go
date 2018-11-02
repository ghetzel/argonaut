package argonaut

/*
Argonaut is a library that makes working with complex command line utilities simpler. Instead of
using complex application logic and string slices for building up command line invocations, you can
model shell commands using Golang structs, then using Argonaut, marshal those structs into valid
strings that can be used to shell out.
*/

import (
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-stockutil/utils"
)

var DefaultArgumentDelimiter = ` `
var DefaultCommandWordSeparator = `-`
var DefaultArgumentKeyPartJoiner = `.`
var DefaultArgumentKeyValueJoiner = DefaultArgumentDelimiter

type CommandName string
type ArgName string

type argonautTag struct {
	Options               []string
	Label                 string
	SkipName              bool
	Required              bool
	Positional            bool
	LongOption            bool
	ForceShort            bool
	SuffixPrevious        bool
	Delimiters            []string
	MutuallyExclusiveWith []string
	KeyPartJoiner         string
	Joiner                string
}

func (self *argonautTag) DelimiterAt(i int) string {
	if len(self.Delimiters) == 0 {
		return DefaultArgumentDelimiter
	} else if i >= len(self.Delimiters) {
		return self.Delimiters[len(self.Delimiters)-1]
	} else {
		return self.Delimiters[i]
	}
}

// Marshals a given struct into a shell-ready command line string.
func Marshal(v interface{}) ([]byte, error) {
	if command, sep, err := generateCommand(v, true); err == nil {
		return []byte(strings.Join(command, sep)), nil
	} else {
		return nil, err
	}
}

// Parses a given struct and returns slice of strings that can be used with os/exec.
func Parse(v interface{}) ([]string, error) {
	if command, _, err := generateCommand(v, true); err == nil {
		return command, err
	} else {
		return nil, err
	}
}

// Parses a given struct and returns slice of strings that can be used with os/exec. Will panic if
// an error occurs.
func MustParse(v interface{}) []string {
	if command, err := Parse(v); err == nil {
		return command
	} else {
		panic(err.Error())
	}
}

// Parses the given value and returns a new *exec.Cmd instance
func Command(v interface{}) (*exec.Cmd, error) {
	var cmd string
	var args []string

	if typeutil.IsEmpty(v) {
		return nil, fmt.Errorf("Cannot parse empty argument into *exec.Cmd")
	}

	if typeutil.IsKind(v, reflect.Struct) {
		if cmdargs, err := Parse(v); err == nil {
			cmd = cmdargs[0]
			args = cmdargs[1:]
		} else {
			return nil, err
		}
	} else if typeutil.IsKind(v, reflect.String) || typeutil.IsArray(v) {
		cmdargs := sliceutil.Stringify(sliceutil.Sliceify(v))

		if len(cmdargs) > 0 {
			cmd = cmdargs[0]
			args = cmdargs[1:]
		} else {
			return nil, fmt.Errorf("Cannot parse empty argument into *exec.Cmd")
		}
	} else {
		return nil, fmt.Errorf("Unexpected type: need struct, string, or []string, got: %T", v)
	}

	return exec.Command(cmd, args...), nil
}

// Parses the given value and returns a new *exec.Cmd instance.  Will panic if an error occurs.
func MustCommand(v interface{}) *exec.Cmd {
	if command, err := Command(v); err == nil {
		return command
	} else {
		panic(err.Error())
	}
}

func generateCommand(v interface{}, toplevel bool) ([]string, string, error) {
	if !typeutil.IsKind(v, reflect.Struct) {
		return nil, ``, fmt.Errorf("struct needed, got %T", v)
	}

	input := structs.New(v)
	command := make([]string, 0)

	if toplevel {
		command = append(command, fmtCommandWord(input.Name()))
	}

	separator := DefaultArgumentDelimiter
	defaults := argonautTag{
		Delimiters:    []string{DefaultArgumentDelimiter},
		KeyPartJoiner: DefaultArgumentKeyPartJoiner,
		Joiner:        DefaultArgumentKeyValueJoiner,
	}

	for _, field := range input.Fields() {
		if !field.IsExported() || field.Tag(`argonaut`) == `-` {
			continue
		}

		if tag, err := parseTag(field.Tag(`argonaut`), &defaults); err == nil {
			var primaryOpt string

			// for marshaling purposes, the option name is determined as:
			//   - the first value of the tag, or, if that's empty...
			//   - the field name formatted to a common default
			if len(tag.Options) > 0 && tag.Options[0] != `` {
				primaryOpt = tag.Options[0]
			} else {
				primaryOpt = fmtCommandWord(field.Name())
			}

			var values []interface{}

			utils.SliceEach(field.Value(), func(i int, value interface{}) error {
				values = append(values, value)
				return nil
			}, reflect.Struct, reflect.Map)

			// arrify and iterate through the field value
			for _, value := range values {
				// CommandName: specifies a named command and options for processing peer fields
				// ---------------------------------------------------------------------------------
				if _, ok := value.(CommandName); ok {
					valueS := fmt.Sprintf("%v", value)

					// specify how the final command should be joined together when marshalling
					if len(tag.Delimiters) > 0 {
						separator = tag.Delimiters[0]
					}

					// if these tags weren't explicitly set, then this is effectively a no-op
					// if they were set, the defaults are updated here to reflect that
					defaults.Delimiters = tag.Delimiters
					defaults.Joiner = tag.Joiner
					defaults.KeyPartJoiner = tag.KeyPartJoiner

					if valueS != `` {
						// prefer value of the field
						command = []string{valueS}

					} else if len(tag.Label) > 0 {
						// fallback to label value
						command = []string{tag.Label}

					} else if primaryOpt != `` {
						// fallback to tag value
						command = []string{primaryOpt}

					} else {
						command = []string{fmtCommandWord(field.Name())}

					}

				} else if _, ok := value.(ArgName); ok {
					// ArgName: specifies a named argument from within a nested struct
					// ---------------------------------------------------------------------------------

					var prefix string

					if tag.ForceShort {
						prefix = `-`
					} else {
						prefix = `--`
					}

					if len(tag.Label) > 0 {
						// prefer label
						command = append(command, prefix+tag.Label)
					} else if primaryOpt != `` {
						command = append(command, prefix+primaryOpt)
					} else {
						command = append(command, prefix+fmtCommandWord(field.Name()))
					}

				} else if typeutil.IsKind(value, reflect.Map) {
					// Maps: get exploded into options
					// ---------------------------------------------------------------------------------

					if err := maputil.Walk(value, func(v interface{}, key []string, isLeaf bool) error {
						if isLeaf {
							var kv string

							if tag.ForceShort {
								kv += `-`
							} else if tag.LongOption {
								kv += `--`
							}

							// only non-nil values expand into [-]key=value arguments
							if v != nil {
								kv += strings.Join(key, tag.KeyPartJoiner)

								if tag.Joiner == separator {
									command = append(command, kv)
									kv = ``
								} else {
									kv += tag.Joiner
								}

								kv += stringutil.MustString(v)
							}

							command = append(command, kv)
						}

						return nil
					}); err != nil {
						return nil, separator, err
					}

				} else if typeutil.IsKind(value, reflect.Struct) {
					// Structs: recurses into this method
					// ---------------------------------------------------------------------------------

					if partial, psep, err := generateCommand(value, false); err == nil {
						// if the separator used in the nested struct matches our own, just tack what
						// came back onto our command stack,
						//
						// otherwise, join using the preferred separator and add it as one big blob
						if psep == separator {
							command = append(command, partial...)
						} else {
							command = append(command, strings.Join(partial, psep))
						}
					} else {
						return nil, separator, err
					}

				} else if tag.SuffixPrevious {
					// SuffixPrevious: modifies the last argument on the command stack with the current value
					// ---------------------------------------------------------------------------------
					if len(command) > 0 && (!typeutil.IsZero(value) || tag.Required) {
						last := command[len(command)-1]

						last += tag.DelimiterAt(0)
						last += stringutil.MustString(value)

						command[len(command)-1] = last
						continue
					}

				} else if tag.Positional {
					// Positional: puts whatever the value is into the command immediately
					// ---------------------------------------------------------------------------------
					command = append(command, sliceutil.Stringify(
						sliceutil.Sliceify(value),
					)...)

					// Scalar Arguments: puts the field name in as the argument name
					//                    boolean fields:  go in as flags (false values are not added)
					//                    everything else: if it has a value or is required, it is added
					// ---------------------------------------------------------------------------------
				} else {
					argName := sliceutil.OrString(primaryOpt, stringutil.Underscore(field.Name()))

					if field.Kind() == reflect.Bool {
						if !typeutil.IsZero(value) {
							command = opt(command, &tag, argName)
						}

					} else if value == nil {
						continue
					} else {
						value = typeutil.ResolveValue(value)

						if !typeutil.IsZero(value) || tag.Required {
							command = opt(command, &tag, argName, sliceutil.Sliceify(value)...)
						}
					}
				}
			}
		} else {
			return nil, separator, err
		}
	}

	return command, separator, nil
}

func fmtCommandWord(in string) string {
	return strings.Replace(
		stringutil.Underscore(in),
		`_`,
		DefaultCommandWordSeparator,
		-1,
	)
}

func opt(command []string, tag *argonautTag, optname string, values ...interface{}) []string {
	argset := []string{}
	prejoin := false

	if !tag.SkipName {
		if tag.LongOption && !tag.ForceShort {
			argset = append(argset, `--`+optname)
			prejoin = true
		} else {
			argset = append(argset, `-`+optname)
		}
	}

	for _, v := range values {
		argset = append(argset, stringutil.MustString(v))
	}

	if prejoin && len(argset) >= 2 {
		command = append(command, argset[0]+tag.Joiner+argset[1])
		command = append(command, argset[2:]...)
	} else {
		command = append(command, argset...)
	}

	return command
}

func parseTag(tag string, defaults *argonautTag) (argonautTag, error) {
	if tag == `` {
		return argonautTag{}, nil
	}

	parts := strings.Split(tag, `,`)

	if len(parts) > 0 {
		argonaut := argonautTag{
			Options:       sliceutil.CompactString(strings.Split(parts[0], `|`)),
			Delimiters:    defaults.Delimiters,
			KeyPartJoiner: defaults.KeyPartJoiner,
			Joiner:        defaults.Joiner,
		}

		for _, tagopt := range parts[1:] {
			optparts := strings.SplitN(tagopt, `=`, 2)

			switch optparts[0] {
			case `required`:
				argonaut.Required = true
			case `positional`:
				argonaut.Positional = true
			case `long`:
				argonaut.LongOption = true
			case `short`:
				argonaut.LongOption = false
				argonaut.ForceShort = true
			case `suffixprev`:
				argonaut.SuffixPrevious = true
			case `skipname`:
				argonaut.SkipName = true
			default:
				if len(optparts) == 1 {
					return argonautTag{}, fmt.Errorf("argonaut tag option %q requires an argument", optparts[0])
				}

				switch optparts[0] {
				case `label`:
					argonaut.Label = optparts[1]
				case `delimiters`, `joiner`, `keyjoiner`:
					v := optparts[1]
					v = strings.TrimPrefix(v, `[`)
					v = strings.TrimSuffix(v, `]`)

					switch optparts[0] {
					case `delimiters`:
						argonaut.Delimiters = strings.Split(v, ``)
					case `joiner`:
						argonaut.Joiner = v
					case `keyjoiner`:
						argonaut.KeyPartJoiner = v
					}
				}
			}
		}

		// if long option wasn't specified, but multiple option names were given
		// assume that the first one is a long option
		if !argonaut.ForceShort && !argonaut.LongOption {
			if len(argonaut.Options) > 1 {
				argonaut.LongOption = true
			}
		}

		return argonaut, nil

	} else {
		return argonautTag{}, nil
	}
}
