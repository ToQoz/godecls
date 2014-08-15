package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	exitCode = 0
	list     = flag.Bool("l", false, "Ouput list of fileames that will be targeted by godecls.")
	noHeader = flag.Bool("h", false, "Never print filenames with output lines.")
	header   = flag.Bool("H", false, "Force print filenames with output lines.")
)

func report(err error) {
	exitCode = 2
	fmt.Fprintln(os.Stderr, err.Error())
}

func usage() {
	exitCode = 2
	fmt.Fprintln(os.Stdout, `godecls lists declarations in files

Usage: godecls [flags] [paths]
`)
	flag.PrintDefaults()
}

func main() {
	defer os.Exit(exitCode)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		if err := proccessFile("<no filename>", os.Stdout, os.Stdin); err != nil {
			report(err)
			return
		}
	} else {
		for i := 0; i < flag.NArg(); i++ {
			fi, err := os.Stat(flag.Arg(i))
			if err != nil {
				report(err)
			}

			if fi.IsDir() {
				if err := walkDir(flag.Arg(i), os.Stdout); err != nil {
					report(err)
				}
			} else {
				*noHeader = true
				if err := proccessFile(flag.Arg(i), os.Stdout, nil); err != nil {
					report(err)
				}
			}
		}
	}

}

func walkDir(filename string, out io.Writer) error {
	return filepath.Walk(filename, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			report(err)
			return nil
		}

		if fi.IsDir() {
			if path == filename {
				return nil
			}

			// Don't walk in sub directories
			return filepath.SkipDir
		}

		if strings.HasSuffix(fi.Name(), ".go") {
			if err := proccessFile(path, out, nil); err != nil {
				report(err)
				return nil
			}
		}

		return nil
	})
}

func proccessFile(filename string, out io.Writer, in io.Reader) error {
	var src []byte

	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	} else {
		*noHeader = true
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	if *list {
		fmt.Fprintln(os.Stderr, filename)
		return nil
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return err
	}

	for _, decl := range file.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			decl := decl.(*ast.GenDecl)
			if decl.Tok == token.IMPORT {
				continue
			}

			for _, spec := range decl.Specs {
				if !*noHeader || *header {
					fmt.Fprint(out, filename+":")
				}
				fprintlnNode(out, fset, decl.Tok.String()+" ", spec)
			}
		case *ast.FuncDecl:
			if !*noHeader || *header {
				fmt.Fprint(out, filename+":")
			}
			fprintlnNode(out, fset, "", decl)
		}

	}

	return nil
}

func fprintlnNode(w io.Writer, fset *token.FileSet, prefix string, node ast.Node) error {
	var buf bytes.Buffer
	fmt.Fprint(&buf, prefix)
	err := printer.Fprint(&buf, fset, node)
	if err != nil {
		return err
	}

	b := buf.Bytes()
	for i, c := range b {
		if c == '\n' {
			if i > 0 && b[i-1] == '{' {
				w.Write([]byte("...}"))
			}
			break
		} else {
			w.Write([]byte{c})
		}
	}
	w.Write([]byte{'\n'})

	return nil
}
