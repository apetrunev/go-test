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
	var path string
	flag.StringVar(&path, "path", "", "path to a file")
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
	source := ast.Source{}
	// read instruction to tokens
	source.Build(&lex)
	source.Print()
}
