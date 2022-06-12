[![Build Status](https://github.com/oligoden/meta/workflows/test%20and%20build/badge.svg)](https://github.com/oligoden/meta/actions?workflow=test%20and%20build)
[![Go Report Card](https://goreportcard.com/badge/github.com/oligoden/meta)](https://goreportcard.com/report/github.com/oligoden/meta)

# Meta Controller

Meta Controller (MC or Meta) is a meta programming tool that allows not only
building of a project's source code but also actively control it. This is
accomplished by an abstract file structure representation of the source code
together with a configuration file (by default `meta.json`). Controlling these
two components provides powerfull control of the resultant source code.

Some features are:
- Switching code between environments
- Transformations and injections using
[Go's templating language](https://pkg.go.dev/text/template)
- Composing and decomposing of files (or fan-in/fan-out)
- Running actions and building files (ordered by a DAG)
- Extend build functionallity to fit various applications 
- and many more...

The broader goal of MC to allow development on a small set of abstact code
while controlling large code bases. In doing so the hope is to improve
programming efficiency and readability while reducing complexity.

> Programs must be written for people to read, and only incidentally
> for machines to execute. (H. Abelson and G. Sussman)

Read more about Meta at [oligoden.com/meta](https://oligoden.com/meta) :south_africa:

---

*MC is part of the [Oligoden](https://oligoden.com) set of tools
and libraries. Checkout some of the other repositories.*

## Table of Contents

* [Getting Started](https://github.com/oligoden/meta#getting-started)
* [Introduction](https://github.com/oligoden/meta#introduction)
  * [Motivation](https://github.com/oligoden/meta#motivation)
  * [Meta Programming](https://github.com/oligoden/meta#meta-programming)
  * [Overview](https://github.com/oligoden/meta#overview)
  * [Usage](https://github.com/oligoden/meta#usage)
* [Installation](https://github.com/oligoden/meta#installation)
* [Configuration Reference](https://github.com/oligoden/meta#meta.json-configuration-reference)
* [Extending to your own builder](https://github.com/oligoden/meta#extending-to-your-own-builder)

## Getting Started

If you have [Go](https://pkg.go.dev/) installed
(and your `~/go/bin` in your PATH env):

```bash
go install github.com/oligoden/meta@latest
```

will install Meta and:

```bash
meta
```

will print the usage.
See [Installing Meta](https://oligoden.com/meta/installing)
for more information. You can now create a config file with:

```bash
echo "{\"name\": \"my-app\"}" > meta.json
mkdir work
```

and build with:

```bash
meta build
```

A file can now be added and Meta configured to include it:

```json
{
  "name": "my-app",
  "directories": {
    "cmd": {
      "dst-ovr":"/.app",
      "options":"output"
      "files": {
        "main.go": {}
      }
    }
  }
}
```

This time run `meta up` to build again. Meta now keeps on running
and will watch for file changes until stopped with `Ctrl-C`.
You should now see a `.app` folder with a `main.go` file in it.
The `cmd/main.go` file is processed(derived) and written to `.app/main.go`.
Derived files be removed by running `meta down`.

Refer to [Getting Started](https://oligoden.com/meta/getting-started)
for more information.

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

## Installation

Download the compressed package from [github.com/oligoden/meta/releases](https://github.com/oligoden/meta/releases)

Optionally, download the `SHA256SUMS` file and run

```bash
sha256sum -c SHA256SUMS 2>&1 | grep OK
```

Extract the package

```bash
#Linux, macOS or FreeBSD
unzip meta-$VERSION-$OS-$ARCH.zip -d /usr/local
```

Add binary to your environment:

```bash
export PATH=$PATH:/usr/local/meta/bin
```

## meta.json Configuration Reference

Refer to [The meta.json Configuration Reference](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md)

## Extending to your own builder

*coming*