[![Build Status](https://github.com/oligoden/meta/workflows/test%20and%20build/badge.svg)](https://github.com/oligoden/meta/actions?workflow=test%20and%20build)
[![Go Report Card](https://goreportcard.com/badge/github.com/oligoden/meta)](https://goreportcard.com/report/github.com/oligoden/meta)

# Meta

Welcome to this meta programming tool where you can build applications by
copying, parsing and bundling files from a meta code directory as described
by a configuration data file. With many additional features, Meta can make
your development and even devops work more efficient.

With Meta you can do the following:
- working on a reduced code base since templates can be shared by multiple files
- improve development speed with the ability to configure applications
- quick switching of the generated source code between different environments

Hope you find it interesting.

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
grown over a process of 3 years.
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
for the templates. Programming is done in a meta code directory on files and
templates that are mapped to the application source code.

A configuration file contains the data used for template parsing and file
placement instructions. Finally, native system commands can also be specified
to transpile or perform other processes on the source code.

A watch mode forms a critical part of the builder. This allows the programmer
to work in the meta code directory instead of the source code directory and
have changes pull through seamlessly into the entire application.

### Overview

The builder uses the `meta.json` configuration file
and the template files (in a meta source folder) to
build the project.

The builder will transfer files from sources to destinations by:
- stepping through directories recursively
- parsing any templates available in the source specified by the "from" key.
- stepping through files
- use file key as source name if "source" field is empty
- use file key as destination name
- create destination file

### Usage

The builder can be executed with:

```bash
meta builder
```

Files that does not exist yet will be created. If files already exists, it will not be replaced. However, the flag `-f` forces the replacement of all files.

Your project might look like this:
```
project-root
|- payload
|  |- file.ext
|
|- meta.json
```
The `payload` folder contains the files and templates that will be used to build files.
The `meta.json` file describes how to build the files and where to place them.

Files will by default be built from the `payload` folder into the `project-root` using `meta.json`. The `-s`, `-d` and `-c` flags can be used to change these defaults.
A `-w` flag can be added to put the builder into
**watch mode** where files will be automatically
updated when the meta code is changed.
Use the `-h` flag for help.

## meta.json Configuration Reference

Refer to [The meta.json Configuration Reference](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md)

## Extending to your own builder

*coming*