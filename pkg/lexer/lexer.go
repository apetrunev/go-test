// lexer
package lexer

import (
	"bufio"
	"errors"
	"io"
	"log"
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
	TokenLess
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
	TokenLess:               "Less",
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
	Reader *bufio.Reader
	Row    int
	Col    int
}

type Tokenizer interface {
	GetToken() Token
	TokenToStr(Token) string
}

// get next token from file
func (l *Lexer) GetToken() Token {
	r, n, err := l.Reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return NoneToken
		}
		log.Fatalf("err:GetToken: %v\n", err)
	}
	switch r {
	case '\n':
		l.Row += 1
		l.Col = 0
		return Token{Type: TokenNewline, Val: "\n", Row: l.Row, Col: l.Col, Len: n}
	case '\t':
		l.Col += 1
		return Token{Type: TokenTab, Val: "\t", Row: l.Row, Col: l.Col, Len: n}
	case ' ':
		l.Col += 1
		return Token{Type: TokenSpace, Val: " ", Row: l.Row, Col: l.Col, Len: n}
	case ':':
		l.Col += 1
		return Token{Type: TokenColon, Val: ":", Row: l.Row, Col: l.Col, Len: n}
	case '#':
		l.Col += 1
		return Token{Type: TokenHash, Val: "#", Row: l.Row, Col: l.Col, Len: n}
	case '=':
		l.Col += 1
		return Token{Type: TokenEqual, Val: "=", Row: l.Row, Col: l.Col, Len: n}
	case '$':
		l.Col += 1
		return Token{Type: TokenDollar, Val: "$", Row: l.Row, Col: l.Col, Len: n}
	case '(':
		l.Col += 1
		return Token{Type: TokenLeftBrace, Val: "(", Row: l.Row, Col: l.Col, Len: n}
	case ')':
		l.Col += 1
		return Token{Type: TokenRightBrace, Val: ")", Row: l.Row, Col: l.Col, Len: n}
	case '.':
		l.Col += 1
		return Token{Type: TokenPeriod, Val: ".", Row: l.Row, Col: l.Col, Len: n}
	case '<':
		l.Col += 1
		return Token{Type: TokenLess, Val: ".", Row: l.Row, Col: l.Col, Len: n}
	case '\'':
		l.Col += 1
		_len := n
		var literal []rune
		for {
			r, n, err := l.Reader.ReadRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return NoneToken
				}
				log.Fatalf("err:GetToken: %s\n", err)
			}
			if r == '\'' || r == '\n' {
				break
			}
			_len += n
			literal = append(literal, r)
		}
		return Token{Type: TokenLiteralQuote, Row: l.Row, Col: l.Col, Len: _len, Val: string(literal)}
	case '"':
		l.Col += 1
		_len := n
		var literal []rune
		for {
			r, n, err := l.Reader.ReadRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return NoneToken
				}
				log.Fatalf("err:GetToken: %s\n", err)
			}
			if r == '"' || r == '\n' {
				break
			}
			_len += n
			literal = append(literal, r)
		}
		return Token{Type: TokenLiteralDoubleQuote, Row: l.Row, Col: l.Col, Len: _len, Val: string(literal)}
	default:
		l.Col += 1
		if unicode.IsLetter(r) {
			var id []rune
			_len := n
			id = append(id, r)
			for {
				// read next rune
				r, n, err := l.Reader.ReadRune()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return NoneToken
					}
					log.Fatalf("err:GetToken: %s\n", err)
				}
				// unread character
				if !unicode.IsLetter(r) {
					// period between two words
					if r != '.' {
						l.Reader.UnreadRune()
						break
					}
				}
				// add rune size to lenth
				_len += n
				id = append(id, r)
			}
			return Token{Type: TokenID, Row: l.Row, Col: l.Col, Len: _len, Val: string(id)}
		} else if unicode.IsDigit(r) {
			var num []rune
			_len := n
			num = append(num, r)
			for {
				r, n, err := l.Reader.ReadRune()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return NoneToken
					}
					log.Fatalf("err:GetToken: %s\n", err)
				}
				// unread character
				if !unicode.IsDigit(r) {
					l.Reader.UnreadRune()
					break
				}
				_len += n
				num = append(num, r)
			}
			return Token{Type: TokenNumber, Row: l.Row, Col: l.Col, Len: _len, Val: string(num)}
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
