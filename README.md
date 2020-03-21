[![Build Status](https://github.com/oligoden/meta/workflows/test%20and%20build/badge.svg)](https://github.com/oligoden/meta/actions?workflow=test%20and%20build)
[![Go Report Card](https://goreportcard.com/badge/github.com/oligoden/meta)](https://goreportcard.com/report/github.com/oligoden/meta)

# Meta (Library an Base Tool)
Build applications by copying, parsing and bundling files from a meta-source directory as described by a configuration data file.

## Meta Programming
Programming on meta files combines the benefits of configuration files and template files to allow rapid development of applications and improved maintainability without sacrificing customisation or adding abstractions. Programming is done in a meta directory on files and templates that are used to generate the regular application source code. A configuration file contains all the data used for template parsing and file placement specification. Finally, commands can be executed on the generated files to compile, transpile or perform any other processes.

Meta programming add the benefits of:
- having a reduced code base since templates can be shared by multiple files,
- improve development speed with the ability to configure applications,
- and allows quick switching of the generated source code between different environments.

A watch mode forms a critical part of the builder. This allows the programmer to work in the meta source directory instead of the regular source directory and having changes pull through seamlessly into the entire application.

## Overview

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

## Using as stand-alone

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

## Implementing your own builder

*coming*

## meta.json Configuration Reference

The config file defines the project at the top level. The project structure is:

```json
{
  "name": "project-name",
  "directories": {},
  "actions": {},
}
```

File structures are specified within an object with key `directories`.
The `directories` object can contain key-value pairs of multiple directories
that can each be viewed as a directory in the project.

```json
{
  "directories": {
    "dir-name": {},
    "dir-name": {},
  }
}
```

The FSs can contain child FSs within an `directories` object, aswell as file objects as key-value pairs under an object with a `files` key.
These are the files that will be built.

```json
{
  "directories": {
    "dir-name": {
      "files": {
        "file-name.ext": {},
        "file-name.ext": {},
      },
      "directories": {
        "dir-name": {
          "files": {}
        }
      }
    }
  }
}
```

By default, a file in the meta directory
will be parsed and written to a file (named with the file key)
and placed under a directory (named with the FS key)
in the project root directory.

```json
{
  "directories": {
    "one": {
      "files": {
        "aaa.ext": {},
        "bbb.ext": {}
      }
    },
    "two": {
      "files": {
        "ccc.ext": {}
      },
      "directories": {
        "six": {
          "files": {
            "jjj.ext": {}
          }
        }
      }
    }
  }
}
```
will build to:

```
./meta/one/aaa.ext -> ./one/aaa.ext
./meta/one/bbb.ext -> ./one/bbb.ext
./meta/two/ccc.ext -> ./two/ccc.ext
./meta/two/six/jjj.ext -> ./two/six/jjj.ext
```

### File Location Modifications

The source and destination paths can be modified with the `from` and `dest` keys. Consider the example:

```json
{
  "directories": {
    "one": {
      "directories": {
        "two": {
          "dest": "D",
          "files": {
            "aaa.ext": {}
          }
        }
      }
    }
  }
}
```

D can be replaced as follows to modify the destination location:
```
D = "" (no effect)
./meta/one/two/aaa.ext -> ./one/two/aaa.ext

D = "sub" (add sub directory)
./meta/one/two/aaa.ext -> ./one/two/sub/aaa.ext

D = "sub/sub" (add sub directories)
./meta/one/two/aaa.ext -> ./one/two/sub/sub/aaa.ext

D = "." (stay in current directory)
./meta/one/two/aaa.ext -> ./one/aaa.ext

D = "sub" (sub directory from current directory)
./meta/one/two/aaa.ext -> ./one/sub/aaa.ext

D = "/" (back to root directory)
./meta/one/two/aaa.ext -> ./aaa.ext

D = "/sub" (sub directory from current directory)
./meta/one/two/aaa.ext -> ./sub/aaa.ext
```

The `from` key can be used in the same way as the `dest` was use above to modify the source location.

### Copying files only

The `copy` key can be used at directory and file level to copy files directly.

```json
{
  "directories": {
    "one": {
      "copy": true,
      "files": {
        "aaa.ext": {},
        "bbb.ext": {}
      }
    },
    "two": {
      "files": {
        "ccc.ext": {},
        "ddd.ext": {"copy": true}
      }
    }
  }
}
```

In the example above, both files in directory `one` are copied but only `ddd.ext`
in directory `two` is copied while `ccc.ext` is parsed as normal.
By default the copy field will be false. If set to true, file parsing will be
skipped.

### Actions

Commands can be executed on the generated files. They are specified in the
`actions` key will typically look like:

```
"actions": {
  "action-a": {
    "pattern": "some-regex",
    "command": ["program", "params", "..."]
  }
}
```

Whenever a filename matches the pattern, the command will be added to the
actions list under the name of the action (`action-a` in the example above).
After all files are generated, the builder will go through this list
and execute the commands. The order is by default non-specific
but can controlled with an optional `dependant` key.

```
"actions": {
  "a": {
    "pattern": "some-regex",
    "command": ["program", "params", "..."],
    "depends-on": ["b", "c"]
  },
  "b": {
    "pattern": "some-regex",
    "command": ["program", "params", "..."],
  },
  "c": {
    "pattern": "some-regex",
    "command": ["program", "params", "..."],
  }
}
```

In the example above action `a` will only be executed after `b` and `c` are executed.

Specifying timeouts are also optional and can be done with the
`timeout` key and an integer (unsigned) giving the time in milliseconds.

```
"actions": {
  "a": {
    "pattern": "some-regex",
    "command": ["program", "params", "..."],
    "timeout": 100
  }
}
```