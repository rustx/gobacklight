# gobacklight

[![Build Status](https://travis-ci.org/rustx/gobacklight.svg?branch=master)](https://travis-ci.org/rustx/gobacklight)

Copyright (C) 2017 - 2018, rustx
This is a free software, see the source for copying conditions. There is NO warranty; not even 
for MERCHANTABILITY or FITNESS for a particular purpose.

## Description

Gobacklight is a program to control backlight controllers for dell XPS 13 ubuntu laptops.

## Features

* Get the current brightness percentage.
* Increment the current brightness with a percentage value between 1 and 10
* Decrement the current brightness with a percentage value between 1 and 10 
* Set the current brightness with a given percentage between 1 and 99.

## Installation

You need to have the [Golang](https://golang.org/doc/install) SDK installed on your system to install this package.

Download the gobacklight sources :

```go get github.com/rustx/gobacklight```

Then, you can install gobacklight package :

```go install github.com/rustx/gobacklight```

Next, you need to add a udev rules file at `/etc/udev/rules.d/90-backlight.rules` allowing the users in the video group to modify the drivers files.

```
ACTION=="add", SUBSYSTEM=="backlight", RUN+="/bin/chgrp video /sys/class/backlight/%k/brightness"
ACTION=="add", SUBSYSTEM=="backlight", RUN+="/bin/chmod g+w /sys/class/backlight/%k/brightness"
```

Export your `GOPATH` into your users `PATH` in /etc/profile file to use gobacklight easier :

`echo "export PATH=$PATH:$GOPATH/bin" >> /etc/profile`

Restart your system to be sure the udev rules are taken into account.

## Usage

To get the help, just use `-h` option :

```
gobacklight -h
Usage:
  gobacklight [OPTIONS]

Application Options:
  -v, --device= brightness device (default: intel_backlight)
  -i, --inc=    increment brightness up to given percentage between [1 -10] (default: nil)
  -d, --dec=    decrement brightness down to percentage between [1 -10] (default: nil)
  -s, --set=    set brightness to given percentage between [1-99] (default: nil)
  -g, --get     get actual brightness percentage

Help Options:
  -h, --help    Show this help message

Examples :
	gobacklight -d intel_backlight -g
	gobacklight -d intel_backlight -i 5
	gobacklight -d intel_backlight -d 5
	gobacklight -d intel_backlight -s 25
```

To use a different `device`, use the `-v` option :

```gobacklight -v "your_device" -g```

To use `get` feature, use -g option :

```gobacklight -g```

To use `inc` or `dec` feature, use -i option with a value between 1 and 10 :

```gobacklight -i 5```
```gobacklight -d 5```

To use `set` feature, use -s option with a value between 1 and 99 :

```gobacklight -s 25```

## Development 

Run tests with coverage :

```
go test --covermode=count -coverprofile=count.out
```

Show coverage per functions :

```
go tool cover -func=count.out
```

Generate html report for coverage :

```
go tool cover -html=count.out
```

