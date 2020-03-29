# The meta.json Configuration Reference

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

## File Location Modifications

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

## Copying files only

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

## Actions

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