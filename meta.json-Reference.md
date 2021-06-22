# The meta.json Configuration Reference

## Index

* [Structure](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md#structure)
  * [File Creation](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md#file-creation)
    * [File Location Modifications](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md#file-location-modifications)
    * [Copying Files Only](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md#copying-files-only)
    * [Including files in files](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md#including-files-in-files-fan-in)
  * [Execs](https://github.com/oligoden/meta/blob/master/meta.json-Reference.md#execs)

## Structure

The config file defines the project at the top level. The project structure is:

```json
{
  "name": "project-name",
  "directories": {},
  "files": {},
  "execs": {},
}
```

### File Creation

File structures are specified within json objects with keys `directories` and `files`.
The `directories` object can contain key-value pairs of multiple directories that
represent directories in the project.

```json
{
  "directories": {
    "dir-name": {},
    "dir-name": {},
  },
  "files": {
    "file-name.ext": {},
    "file-name.ext": {},
  }
}
```

The `directories` can contain child `directories` objects, as well as `files` objects as key-value pairs.
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

By default, a file in the source directory
will be parsed and written to a file (named as the file key)
and placed in the destination directory (with the same name).

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
      "directories": {
        "look": {
          "files": {
            "cat.ext": {}
          }
        }
      },
      "files": {
        "ccc.ext": {}
      }
    }
  },
  "files": {
    "ddd.ext": {},
    "eee.ext": {}
  }
}
```
will build to:

```
./one/aaa.ext      -> ./dst/one/aaa.ext
./one/bbb.ext      -> ./dst/one/bbb.ext
./two/ccc.ext      -> ./dst/two/ccc.ext
./two/look/cat.ext -> ./dst/two/look/cat.ext
./ddd.ext          -> ./dst/ddd.ext
./eee.ext          -> ./dst/eee.ext
```

#### File Location Modifications

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
D = empty (no effect)
./meta/one/two/aaa.ext -> ./one/two/aaa.ext

D = sub (add sub directory)
./meta/one/two/aaa.ext -> ./one/two/sub/aaa.ext

D = sub/sub (add sub/sub directories)
./meta/one/two/aaa.ext -> ./one/two/sub/sub/aaa.ext

D = . (stay in current directory)
./meta/one/two/aaa.ext -> ./one/aaa.ext

D = ./sub (sub directory from current directory)
./meta/one/two/aaa.ext -> ./one/sub/aaa.ext

D = / (back to root directory)
./meta/one/two/aaa.ext -> ./aaa.ext

D = /sub (sub directory from root directory)
./meta/one/two/aaa.ext -> ./sub/aaa.ext
```

The `from` key can be used in the same way as the `dest` was use above to modify the source location.

#### Copying Files Only

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

#### Including files in files (fan-in)

Files can be included into other files. This is also called fan-in sinse one
file is built from multiple sources. To do this, use the `templates` key in
the file to specify the files.

```json
{
  "directories": {
    "one": {
      "files": {
        "aaa.ext": {
          "templates": [
            "one/bbb.ext",
            "two/ccc.ext"
          ]
        },
        "bbb.ext": {}
      }
    },
    "two": {
      "files": {
        "ccc.ext": {}
      }
    }
  }
}
```

Files `bbb.ext` and `ccc.ext` can then be included into `aaa.ext`.

```none
{{template "bbb.ext"}}
{{template "ccc.ext"}}
```

### Execs

Commands can be executed on the generated files. They are specified in the
`execs` key will typically look like:

```
"execs": {
  "exec-a": {
    "cmd": ["program", "params", "..."]
  }
}
```

Whenever a node is updated, the commands linked to the node and all the parent
nodes will be executed in the order specified by the tree. Specifying timeouts
are also optional and can be done with the `timeout` key and an integer
(unsigned) giving the time in milliseconds.

```
"execs": {
  "a": {
    "cmd": ["program", "params", "..."],
    "timeout": 100
  }
}
```