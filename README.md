[![Build Status](https://github.com/oligoden/meta/workflows/test%20and%20build/badge.svg)](https://github.com/oligoden/meta/actions?workflow=test%20and%20build)
[![Go Report Card](https://goreportcard.com/badge/github.com/oligoden/meta)](https://goreportcard.com/report/github.com/oligoden/meta)

# Meta

Welcome to this meta programming tool where you can develop applications by
copying, parsing and bundling files from a meta (control) code directory and a
meta configuration file. With many additional features, Meta aims to make your
development and deployment work more efficient.

Meta uses template based meta programming that allows you to:

- make **rapid** code changes by working a control code base and configuration
file. Files can be composed in powerfull ways so that a change in one file of
the control code can update data structures, variables, etc. in multiple files
in the source code.

- **simplify** the code you need read or develop on by only working on reduced
control code or the configuration file. The template based meta programming and
file composition approach makes it easier to keep your code DRY without complex
syntax that use highly specialised techniques.

- **automate** transpiling files, compilation, .etc by running external tools
as soon as changes are detected.

Hope you find it interesting. :smile:

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

```
go install github.com/oligoden/meta
meta build
```

## Introduction

### Motivation

In an attempt to come up with techniques to increase application development
rates without lossing accessibility to source code, Meta was developed and
matured since 2017.

A tradeoff exists with simply adding abstractions. By definition the programmer
gets removed from the under lying source code and losses "power" in exchange
for ease.
Meta follows an orhorgonal approach by giving you more perspective over your
source code from where you can better control and monitor it, instead of hiding
it.
How do we accomplish all that? By doing programming of programming or in short,
meta programming.

### Meta Programming - our take on it
There are many ways to apply meta programming. We use a template based approach
and a configuration file to control files and template mapping and provide data
for the templates. Programming is done in a control code directory on files and
templates that are mapped to the application source code.

A configuration file contains the data used for template parsing and file
placement instructions. Finally, native system commands can also be specified
to transpile or perform other external processes on the source code.

A watch mode forms a critical part of the builder. This allows the programmer
to work in the control code directory instead of the full source code and
have changes pull through seamlessly into the entire application.

### Overview

Meta is a stand alone application that you pre-install run in your project.
Once installed you `cd` to your project folder and run the Meta builder with
`meta build`.

The builder uses the `meta.json` configuration file and template or regular
files (in a control code directory traditionally named `work`) to build the
project. Here is an configuration file example:

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

It starts with a project name and is followed with the `directories` key that
specifies a file structure of files that will be mapped to the source code.
It is then followed with the `execs` key that specifies external commands that
must run once the source code is created.
The example above will take the file `work/app/main.go`, copy or parse it and
create a file `app/main.go`. The Golang build tool will then be invoked to
build a binary.

### Usage

The builder can be executed with:

```bash
meta build
```

Files that does not exist yet will be created. If files already exists, it will not be replaced. However, the flag `-f` forces the replacement of all files.

Your project might look like this:
```
project-root
|- work
|  |- file.ext
|
|- meta.json
```
The `work` folder contains the files and templates that will be used to build files.
The `meta.json` file describes how to build the files and where to place them.

Files will by default be built from the `work` folder into the `project-root` using `meta.json`. The `-s`, `-d` and `-c` flags can be used to change these defaults.
A `-w` flag can be added to put the builder into
**watch mode** where files will be automatically
updated when the meta code is changed.
Use the `-h` flag for help.

## meta.json Configuration Reference

Refer to [The meta.json Configuration Reference](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md)

## Extending to your own builder

*coming*