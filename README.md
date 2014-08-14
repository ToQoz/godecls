# Godecls

```
$ ./godecls --help
godecls lists declarations in files

Usage: godecls [flags] [paths]

  -h=false: Never print filenames with output lines.
  -l=false: Ouput list of fileames that will be targeted by godecls.
```

## Usage

```
$ vi main.go
$ cat main.go
package main

import (
	"os"
)

var (
	exitCode = 0
)

func main() {
	defer os.Exit(exitCode)
}

func runMain() (err error) {
    return nil
}

$ godecls main.go
var exitCode = 0
func main() {...}
func runMain() (err error) {...}
```

## See Also

- https://github.com/soh335/unite-outline-go
