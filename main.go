// parse makefile
package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/apetrunev/go-test/pkg/ast"
	"github.com/apetrunev/go-test/pkg/lexer"
)

func main() {
	var path, out string
	flag.StringVar(&path, "path", "", "path to a file")
	flag.StringVar(&out, "out", "", "output file")
	flag.Parse()
	if path == "" {
		log.Fatalf("err: no file to parse\n")
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	lex := lexer.Lexer{Reader: reader}
	var o *os.File
	if out != "" {
		o, err = os.OpenFile(out, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	} else {
		o = os.Stdout
	}
	defer o.Close()
	source := ast.Source{}
	// read instruction to tokens
	source.Build(&lex)
	source.Expand()
	source.Print(o)
}
