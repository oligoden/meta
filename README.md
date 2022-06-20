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

> The significant problems we face cannot be solved
> by the same level of thinking that created them.
Albert Einstein

Read more about Meta at [oligoden.com/meta](https://oligoden.com/meta) :south_africa:

---

*MC is part of the [Oligoden](https://oligoden.com) set of tools
and libraries. Checkout some of the other repositories.*

---

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
  "dirs": {
    "cmd": {
      "dest":"/.app",
      "options":"output"
      "files": {
        "main.go": {}
      }
    }
  }
}
```

This time run `meta up`. Meta builds again and keeps on running
to watch for file changes until stopped with `Ctrl-C`.
You should now see a `.app` folder with a `main.go` file in it.
The `cmd/main.go` file is processed(derived) and written to `.app/main.go`.
Derived files be removed by running `meta down`.

Refer to [Getting Started](https://oligoden.com/meta/getting-started)
for more information.

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