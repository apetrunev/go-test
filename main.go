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
	TokenQuote
	TokenDoubleQuote
	TokenLiteralQuote
	TokenLiteralDoubleQuote
)

var TokenStr = map[TokenType]string{
	TokenNone:               "None",
	TokenID:                 "ID",
	TokenNumber:             "Number",
	TokenColon:              "Colon",
	TokenTab:                "Tab",
	TokenSpace:              "Space",
	TokenNewline:            "Newline",
	TokenHash:               "Hash",
	TokenEqual:              "Equal",
	TokenDollar:             "Dollar",
	TokenLeftBrace:          "LeftBrace",
	TokenRightBrace:         "RightBrace",
	TokenPeriod:             "Period",
	TokenQuote:              "Quote",
	TokenDoubleQuote:        "DoubleQuote",
	TokenLiteralQuote:       "LiteralQuote",
	TokenLiteralDoubleQuote: "LiteralDoubleQuote",
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

type Tokenizer interface {
	GetToken() Token
	TokenToStr(Token) string
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
		return Token{Type: TokenNewline, Val: "\n", Row: l.row, Col: l.col, Len: n}
	case '\t':
		l.col += 1
		return Token{Type: TokenTab, Val: "\t", Row: l.row, Col: l.col, Len: n}
	case ' ':
		l.col += 1
		return Token{Type: TokenSpace, Val: " ", Row: l.row, Col: l.col, Len: n}
	case ':':
		l.col += 1
		return Token{Type: TokenColon, Val: ":", Row: l.row, Col: l.col, Len: n}
	case '#':
		l.col += 1
		return Token{Type: TokenHash, Val: "#", Row: l.row, Col: l.col, Len: n}
	case '=':
		l.col += 1
		return Token{Type: TokenEqual, Val: "=", Row: l.row, Col: l.col, Len: n}
	case '$':
		l.col += 1
		return Token{Type: TokenDollar, Val: "$", Row: l.row, Col: l.col, Len: n}
	case '(':
		l.col += 1
		return Token{Type: TokenLeftBrace, Val: "(", Row: l.row, Col: l.col, Len: n}
	case ')':
		l.col += 1
		return Token{Type: TokenRightBrace, Val: ")", Row: l.row, Col: l.col, Len: n}
	case '.':
		l.col += 1
		return Token{Type: TokenPeriod, Val: ".", Row: l.row, Col: l.col, Len: n}
	case '\'':
		l.col += 1
		_len := n
		var literal []rune
		for {
			r, n, err := l.reader.ReadRune()
			if err != nil {
				log.Fatalf("err:GetToken: %s\n", err)
			}
			if r == '\'' || r == '\n' {
				break
			}
			_len += n
			literal = append(literal, r)
		}
		return Token{Type: TokenLiteralQuote, Row: l.row, Col: l.col, Len: _len, Val: string(literal)}
	case '"':
		l.col += 1
		_len := n
		var literal []rune
		for {
			r, n, err := l.reader.ReadRune()
			if err != nil {
				log.Fatalf("err:GetToken: %s\n", err)
			}
			if r == '"' || r == '\n' {
				break
			}
			_len += n
			literal = append(literal, r)
		}
		return Token{Type: TokenLiteralDoubleQuote, Row: l.row, Col: l.col, Len: _len, Val: string(literal)}
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
					// period between two words
					if r != '.' {
						l.reader.UnreadRune()
						break
					}
				}
				// add rune size to lenth
				_len += n
				id = append(id, r)
			}
			return Token{Type: TokenID, Row: l.row, Col: l.col, Len: _len, Val: string(id)}
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
			return Token{Type: TokenNumber, Row: l.row, Col: l.col, Len: _len, Val: string(num)}
		}
		return NoneToken
	}
	return NoneToken
}

func (l *Lexer) TokenToStr(t Token) string {
	s, ok := TokenStr[t.Type]
	if !ok {
		log.Fatalf("err:TokenToStr: unknown token %v\n", t.Type)
	}
	return s
}

type AstNodeType int

const (
	AstNodeTarget = iota
	AstNodeAssignment
)

type Ast interface {
	Type() AstNodeType
}

type Node struct {
	Type AstNodeType
}

type Expr interface {
	Value() string
}

type ExprID struct {
	ID string
}

func (e *ExprID) Value() string {
	return ""
}

type ExprVar struct {
	ExprID
	Val string
}

func (e *ExprVar) Value() string {
	return e.Val
}

type NodeTarget struct {
	Node
	ID            Expr
	Prerequisites []Expr
	Recipe        []Expr
}

type Source struct {
	Tree    []Ast
	Symbols map[string]*ExprVar
}

// stmt_target -> list_pre (newline list_recipe)
// list_pre -> list_pre expr_term
// expr_term -> $(term) | term
// term -> id | literal
// list_recipe -> list_recipe tab expr_cmd newline
// expr_cmd -> expr_cmd expr_term

func (s *Source) prerequisites(lex Tokenizer, n *NodeTarget) {

}

func (s *Source) exprVar(lex Tokenizer, t Token) Expr {
	t = lex.GetToken()
	switch t.Type {
	case TokenLeftBrace:
		for t.Type != TokenRightBrace || t.Type != TokenNewline {
			t = lex.GetToken()
			switch t.Type {
			case TokenID:
				// get from symbol table
				term := ExprVar{ExprID: ExprID{ID: t.Val}}
				return &term
			default:
				log.Fatalf("err:exprVar: expected id but got %s\n", lex.TokenToStr(t))
			}
		}
	default:
		log.Fatalf("err:exprVar: expected ( but found %s\n", lex.TokenToStr(t))
	}
	return nil
}

func (s *Source) exprTerm(lex Tokenizer, t Token) Expr {
	switch t.Type {
	case TokenID:
		term := ExprID{ID: t.Val}
		return &term
	case TokenDollar:
		term := s.exprVar(lex, t)
		return term
	}
	return nil
}

func (s *Source) Build(lex Tokenizer) {
	for t := lex.GetToken(); t.Type != TokenNone; t = lex.GetToken() {
		switch t.Type {
		case TokenNewline:
			continue
		case TokenID, TokenDollar:
			expr := s.exprTerm(lex, t)
			fmt.Printf("debug: %v\n", expr)
			continue
		default:
			log.Printf("info:Build: %v\n", t)
		}
	}
}

func (s *Source) skipSpaces(lex Tokenizer) Token {
	t := lex.GetToken()
	for t.Type == TokenSpace {
		t = lex.GetToken()
	}
	return t
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
	source := Source{}
	// read instruction to tokens
	source.Build(&lex)
}
