[![Build Status](https://github.com/oligoden/meta/workflows/test%20and%20build/badge.svg)](https://github.com/oligoden/meta/actions?workflow=test%20and%20build)
[![Go Report Card](https://goreportcard.com/badge/github.com/oligoden/meta)](https://goreportcard.com/report/github.com/oligoden/meta)

# Meta

Welcome to a meta programming tool where you can develop applications by
copying, parsing and bundling files from a meta (control) code directory and
configuration file. With many additional features, Meta aims to make your
development and deployment work more efficient.

Meta uses template based meta programming that allows you to:

- make **rapid** code changes by working on a control code base and configuration
file. Files can be composed in powerful ways so that a change in one file of
the control code can update data structures, variables, etc. in multiple files
in the source code.

- **simplify** the code you read or develop by only working on reduced
control code or the configuration file. The template based meta programming and
file composition approach makes it easier to keep your code DRY without complex
syntax that uses highly specialised techniques.

- **automate** transpiling files, compilation, etc. by running external tools
as soon as changes are detected.

Hope you find it interesting. :smiley: :south_africa:

## Table of Contents

* [Quick Start](https://github.com/oligoden/meta#quick-start)
* [Introduction](https://github.com/oligoden/meta#introduction)
  * [Motivation](https://github.com/oligoden/meta#motivation)
  * [Meta Programming](https://github.com/oligoden/meta#meta-programming)
  * [Overview](https://github.com/oligoden/meta#overview)
  * [Usage](https://github.com/oligoden/meta#usage)
* [Configuration Reference](https://github.com/oligoden/meta#meta.json-configuration-reference)
* [Extending to your own builder](https://github.com/oligoden/meta#extending-to-your-own-builder)

## Quick Start

If you have Golang installed (and your `~/go/bin` in your PATH env):

```bash
go install github.com/oligoden/meta
```

should install Meta and:

```bash
meta
```

will print the usage. You can now create a config file and a work directory with:

```bash
echo "{}" > test.file
mkdir work
```

and build with:

```bash
meta build
```

## Introduction

### Motivation

In an attempt to come up with techniques to increase application development
rates, without lossing accessibility to source code, Meta has been developed and
matured from 2017.

A trade-off exists by simply adding abstractions. By definition the programmer
gets removed from the under lying source code and loses "power" in exchange
for ease.
Meta follows an orthogonal approach by giving you more perspective over your
source code from where you can better control and monitor it, instead of hiding
it.
How do we accomplish all that? By doing programming of programming or in short,
meta programming.

### Meta Programming - our take on it
There are many ways to apply meta programming. We use a template based approach
and a configuration file to control file mapping and provide data
for templates. Programming is done in a control code directory on files and
templates that are mapped to the application source code.

A configuration file contains the data used for template parsing and file
placement instructions. Finally, native system commands can also be specified
to transpile or perform other external processes on the source code.

A watch mode forms a critical part of the builder. This allows the programmer
to work in the control code directory instead of the full source code and
have changes pull through seamlessly into the entire application.

### Overview

Meta is a stand alone application that is pre-installed and run in your project.
Once installed, `cd` to your project folder and run the Meta builder with
`meta build`.

The builder uses the `meta.json` configuration file and template or regular
files (in a control code directory traditionally named `work`) to build the
project. Here is an example of a configuration file:

```json
{
  "name": "my-app",
  "directories": {
    "app": {
      "files": {
        "main.go": {}
      }
    }
  },
  "execs": {
    "build": {
      "cmd": ["go", "build", "app.main.go"]
    }
  }
}
```

It starts with a project name and is followed by the `directories` key which
specifies a file structure of files that will be mapped to the source code.
It is then followed by the `execs` key that specifies external commands that
must run once the source code is created.
The example above will take the file `work/app/main.go`, copy or parse it and
create a file `app/main.go`. The Golang build tool will then be invoked to
build a binary.

### Usage

The builder can be executed with:

```bash
meta build
```

Files that do not exist yet will be created. If files already exist, it will not be replaced. However, the flag `-f` forces the replacement of all files.

Your project might look like this:
```
project-root
|- work
|  |- file.ext
|
|- meta.json
```
The `work` folder contains the files and templates that will be used to build files.
The `meta.json` file specifies how to build the files and where to place them.

Files will, by default, be built from the `work` folder into the `project-root` using `meta.json`. The `-s`, `-d` and `-c` flags can be used to alter these defaults.
A `-w` flag can be added to put the builder into
**watch mode** where files will be automatically
updated when the meta code is changed.
Use the `-h` flag for help.

## meta.json Configuration Reference

Refer to [The meta.json Configuration Reference](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md)

## Extending to your own builder

*coming*