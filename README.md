Gowatch
========

Watch for source code changes, then automatically rebuild and run your Golang
project.

This project is currently working on **OSX only**. Windows, Linux, and BSD
development is planned.

## Usage

First, install `gowatch` with

```
$ go get -u github.com/mandala/gowatch
```

Then, run `gowatch` from your project directory.

```
$ gowatch
```

Gowatch will build and run your project and waits for source code changes.

## Automatic Project Reload

You can also watch for folder changes and automatically reload the application
without rebuilding the Golang project. It usually useful for watching resources
folder changes in web application project.

To use the automatic project reload, pass `-r` option after `gowatch`.

```
$ gowatch -r resources/static -r resources/views
```

## Bugs and Feature Requests

Please report them on <https://github.com/mandala/gowatch/issues>.

## Copyright Notice

Copyright (c) 2017 Fadhli Dzil Ikram. All rights reserved.

Use of source code is governed by a MIT license that can be found in the
LICENSE file.
