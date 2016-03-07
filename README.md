# Gobserve

An automatic command launcher using file notification changes.

The main goal is to be able to launch a command when a file changes in a directory. For example, to relaunch a Golang service when you change source, or to launch "make" when you change a source file.

Gobserve uses yaml file to configure

- directories to watch
- ignore list
- command to launch


# Launching

Simply call `gobserve` in a directory. You may use a configuration file (yaml format) to configure the watcher.

# YAML conf

An example to relaunch a golang server:

```yaml
command: go run main.go
watch:
- ./
ignore:
- "*.*~"
- "*/.git"
```

An example to call "make" when a ".c" file changes:


```yaml
command: make
watch:
- "*/*.c"
ignore:
- "*.*~"
- "*/.git"
```


**Note:** "`*.*~`' should be default. This represents temporary files that are created by vim, gedit, or other text editors.

