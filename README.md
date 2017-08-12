# Argonaut [![GoDoc](https://godoc.org/github.com/ghetzel/argonaut?status.svg)](https://godoc.org/github.com/ghetzel/argonaut)

## Overview

Argonaut is a library that makes working with complex command line utilities simpler.  Instead of using complex application logic and string slices for building up command line invocations, you can model shell commands using Golang structs, then using Argonaut, marshal those structs into valid strings that can be used to shell out.
