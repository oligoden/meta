# The meta.json Configuration Reference

## Table of Contents

* [Structure](https://github.com/oligoden/meta.json-configuration-reference#structure)
  * [File Creation](https://github.com/oligoden/meta.json-configuration-reference#file-creation)
    * [File Location Modifications](https://github.com/oligoden/meta.json-configuration-reference#file-location-modifications)
    * [Copying Files Only](https://github.com/oligoden/meta.json-configuration-reference#copying-files-only)
  * [Execs](https://github.com/oligoden/meta.json-configuration-reference#execs)
* [Installation](https://github.com/oligoden/meta#installation)
* [Configuration Reference](https://github.com/oligoden/meta#meta.json-configuration-reference)
* [Extending to your own builder](https://github.com/oligoden/meta#extending-to-your-own-builder)

## Structure

The config file defines the project at the top level. The project structure is:

```json
{
  "name": "project-name",
  "directories": {},
  "execs": {},
}
```

### File Creation

File structures are specified within an object with key `directories`.
The `directories` object can contain key-value pairs of multiple directories
that that represent directories in the project.

```json
{
  "directories": {
    "dir-name": {},
    "dir-name": {},
  }
}
```

The `directories` can contain child `directories` objects, aswell as `files` objects as key-value pairs.
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

By default, a file in the work directory
will be parsed and written to a file (named with the file key)
and placed under a directory (named with the directory key)
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
nodes will be executed in the order specified by the tree.

Specifying timeouts are also optional and can be done with the
`timeout` key and an integer (unsigned) giving the time in milliseconds.

```
"execs": {
  "a": {
    "cmd": ["program", "params", "..."],
    "timeout": 100
  }
}
```