# Simple Compiler

## How to run

First of all, download the source code of the compiler, then run the command: 

```bash
go build cmd/main.go
```
and after that, run the command:
```bash
./main <file_to_parse>
```
to build the compiler and generate the binary or:

```bash
go run xmd/main.go <file_to_parse>
```

Or if you're want to install the compiler at a linux machine, just run 
``` bash
go install cmd/main.go
```
