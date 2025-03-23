# ProdBlockTUI
A customizable program blocker and process killer

## Usage:
As there's no persistent configuration as of yet,
configuration must be done in the source code.

#### Steps:
Locate the Ruleset Rules[] literal 

```var Ruleset = []Rules{}```

Modify the struct slice to the configuration you desire

In your shell:

$ go run . 

OR

$ go build

This will be fixed eventually.

## Build:
Install go

Clone repo

In your shell:

$ go mod tidy 

$ go build 

## TODO:
Refactor

Document source code

Enforce validation of input data at point of configuration

Persistent configuration

Encryption

Non-destructive obfuscation of files

Smarter program lookups (use fuzzy finding to match executable names in subdirectories)

Improve UX

Autorun on executable startup

GUI + CLI + wrapper (curse forced SUBSYSTEM at linkage)
