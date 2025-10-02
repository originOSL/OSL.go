# OSL.go

A go based compiler for osl, intended to allow me to write fast and powerful servers in osl-like syntax. This is unlikely to have complete 1:1 parity with originOS's OSL but it should get close enough!


## Installation

```bash
git clone https://github.com/originOSL/OSL.go
cd OSL.go
go build
```

Then you can run sudo ./osl setup to install the compiler to /usr/local/bin/osl

## Usage

```bash
osl compile main.osl
```

This will compile main.osl and output ./main

Just run osl on its own to get a list of commands

```bash
osl
```
