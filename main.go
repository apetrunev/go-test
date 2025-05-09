// parse makefile
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode"
)

type TokenType int

const (
	TokenNone TokenType = iota
	TokenID
	TokenNumber
	TokenColon
	TokenTab
	TokenSpace
	TokenNewline
	TokenHash
	TokenEqual
	TokenDollar
	TokenLeftBrace
	TokenRightBrace
	TokenPeriod
)

var TokenStr = map[TokenType]string{
	TokenNone:       "None",
	TokenID:         "ID",
	TokenNumber:     "Number",
	TokenColon:      "Colon",
	TokenTab:        "Tab",
	TokenSpace:      "Space",
	TokenNewline:    "Newline",
	TokenHash:       "Hash",
	TokenEqual:      "Equal",
	TokenDollar:     "Dollar",
	TokenLeftBrace:  "LeftBrace",
	TokenRightBrace: "RightBrace",
	TokenPeriod:     "Period",
}

type Token struct {
	Type TokenType
	Val  string
	Len  int
	Row  int
	Col  int
}

// predefined tokens
var NoneToken = Token{Type: TokenNone, Row: -1, Col: -1, Len: -1}

type Lexer struct {
	reader *bufio.Reader
	row    int
	col    int
}

// get next token from file
func (l *Lexer) GetToken() Token {
	r, n, err := l.reader.ReadRune()
	if err != nil {
		log.Fatalf("err:GetToken: %v\n", err)
	}
	switch r {
	case '\n':
		l.row += 1
		l.col = 0
		return Token{Type: TokenNewline, Row: l.row, Col: l.col, Len: n}
	case '\t':
		l.col += 1
		return Token{Type: TokenTab, Row: l.row, Col: l.col, Len: n}
	case ' ':
		l.col += 1
		return Token{Type: TokenSpace, Row: l.row, Col: l.col, Len: n}
	case ':':
		l.col += 1
		return Token{Type: TokenColon, Row: l.row, Col: l.col, Len: n}
	case '#':
		l.col += 1
		return Token{Type: TokenHash, Row: l.row, Col: l.col, Len: n}
	case '=':
		l.col += 1
		return Token{Type: TokenEqual, Row: l.row, Col: l.col, Len: n}
	case '$':
		l.col += 1
		return Token{Type: TokenDollar, Row: l.row, Col: l.col, Len: n}
	case '(':
		l.col += 1
		return Token{Type: TokenLeftBrace, Row: l.row, Col: l.col, Len: n}
	case ')':
		l.col += 1
		return Token{Type: TokenRightBrace, Row: l.row, Col: l.col, Len: n}
	case '.':
		l.col += 1
		return Token{Type: TokenPeriod, Row: l.row, Col: l.col, Len: n}
	default:
		l.col += 1
		if unicode.IsLetter(r) {
			var id []rune
			_len := n
			id = append(id, r)
			for {
				// read next rune
				r, n, err := l.reader.ReadRune()
				if err != nil {
					log.Fatalf("err:GetToken: %s\n", err)
				}
				// unread character
				if !unicode.IsLetter(r) {
					l.reader.UnreadRune()
					break
				}
				// add rune size to lenth
				_len += n
				id = append(id, r)
			}
			return Token{Type: TokenID, Row: l.row, Col: l.col, Len: _len}
		} else if unicode.IsDigit(r) {
			var num []rune
			_len := n
			num = append(num, r)
			for {
				r, n, err := l.reader.ReadRune()
				if err != nil {
					log.Fatalf("err:GetToken: %s\n", err)
				}
				// unread character
				if !unicode.IsDigit(r) {
					l.reader.UnreadRune()
					break
				}
				_len += n
				num = append(num, r)
			}
			return Token{Type: TokenNumber, Row: l.row, Col: l.col, Len: _len}
		}
		return NoneToken
	}
	return NoneToken
}

func (l *Lexer) TokenToStr(t TokenType) string {
	s, ok := TokenStr[t]
	if !ok {
		log.Fatalf("err:TokenToStr: unknown token %v\n", t)
	}
	return s
}

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
	lex := Lexer{reader: reader}
	// read instruction to tokens
	for t := lex.GetToken(); t.Type != TokenNone; t = lex.GetToken() {
		fmt.Printf("%s\n", lex.TokenToStr(t.Type))
	}

}
