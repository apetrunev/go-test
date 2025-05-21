// ast
package ast

import (
	"log"
	"strings"

	"github.com/apetrunev/go-test/pkg/lexer"
)

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

const (
	ExprAllPrerequisites = "$^"
)

type Expr interface {
	Value() string
}

type ExprID struct {
	ID string
}

func (e *ExprID) Value() string {
	return e.ID
}

type ExprVar struct {
	ExprID
	Val string
}

func (e *ExprVar) Value() string {
	return e.Val
}

type ExprCmd struct {
	Terms []Expr
}

func (e *ExprCmd) Value() string {
	var cmd []string
	for _, term := range e.Terms {
		cmd = append(cmd, term.Value())
	}
	return strings.Join(cmd, " ")
}

type NodeTarget struct {
	Node
	ID            Expr
	Prerequisites []Expr
	Recipe        []Expr
}

func (n *NodeTarget) Type() AstNodeType {
	return n.Node.Type
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

func (s *Source) recipe(lex lexer.Tokenizer) Expr {
	// current token is tab
	var cmd ExprCmd = ExprCmd{}
	for t := lex.GetToken(); t.Type != lexer.TokenNewline; t = lex.GetToken() {
		switch t.Type {
		case lexer.TokenID, lexer.TokenDollar:
			term := s.exprTerm(lex, t)
			cmd.Terms = append(cmd.Terms, term)
		case lexer.TokenSpace:
			continue
		default:
			log.Fatalf("err:recipe: %v\n", lex.TokenToStr(t))
		}
	}
	return &cmd
}

func (s *Source) prerequisites(lex lexer.Tokenizer) []Expr {
	// current toke is TokenColon
	var deps []Expr
	for t := lex.GetToken(); t.Type != lexer.TokenNewline; t = lex.GetToken() {
		switch t.Type {
		case lexer.TokenID, lexer.TokenDollar:
			term := s.exprTerm(lex, t)
			deps = append(deps, term)
		case lexer.TokenSpace:
			continue
		default:
			log.Fatalf("err:prerequisites: %v\n", lex.TokenToStr(t))
		}
	}
	return deps
}

func (s *Source) exprVar(lex lexer.Tokenizer, t lexer.Token) Expr {
	// current token is TokenDollar
	t = lex.GetToken()
	switch t.Type {
	case lexer.TokenLeftBrace:
		for t.Type != lexer.TokenRightBrace || t.Type != lexer.TokenNewline || t.Type != lexer.TokenNone {
			t = lex.GetToken()
			switch t.Type {
			case lexer.TokenID:
				// get from symbol table
				term := ExprVar{ExprID: ExprID{ID: t.Val}}
				return &term
			default:
				log.Fatalf("err:exprVar:1 expected id but got %s\n", lex.TokenToStr(t))
			}
		}
	case lexer.TokenLess:
		// special var $<
		term := ExprVar{ExprID: ExprID{ID: ExprAllPrerequisites}, Val: ExprAllPrerequisites}
		return &term
	default:
		log.Fatalf("err:exprVar:2 expected ( but found %s\n", lex.TokenToStr(t))
	}
	return nil
}

func (s *Source) exprTerm(lex lexer.Tokenizer, t lexer.Token) Expr {
	switch t.Type {
	case lexer.TokenID:
		term := ExprID{ID: t.Val}
		return &term
	case lexer.TokenDollar:
		term := s.exprVar(lex, t)
		return term
	}
	return nil
}

func (s *Source) target(lex lexer.Tokenizer, lhs Expr) {
	// current token is TokenColon
	switch _lhs := lhs.(type) {
	case *ExprID, *ExprVar:
		deps := s.prerequisites(lex)
		var r []Expr
		for t := lex.GetToken(); t.Type == lexer.TokenTab; t = lex.GetToken() {
			switch t.Type {
			case lexer.TokenTab:
				cmd := s.recipe(lex)
				r = append(r, cmd)
			default:
				log.Fatalf("err:target: expected TAB but found %s\n", lex.TokenToStr(t))
			}
		}
		var tNode NodeTarget
		tNode.Node.Type = AstNodeTarget
		tNode.ID = lhs
		tNode.Prerequisites = deps
		tNode.Recipe = r
		s.Tree = append(s.Tree, &tNode)
	default:
		log.Fatalf("err:target: expected ID or VAR but found %v\n", _lhs)
	}
}

func (s *Source) assignment(lex lexer.Tokenizer, lhs Expr) {

}

func (s *Source) Build(lex lexer.Tokenizer) {
	for t := lex.GetToken(); t.Type != lexer.TokenNone; t = lex.GetToken() {
		switch t.Type {
		case lexer.TokenNewline:
			continue
		case lexer.TokenID, lexer.TokenDollar:
			// lfs
			expr := s.exprTerm(lex, t)
			tt := s.skipSpaces(lex)
			switch tt.Type {
			case lexer.TokenColon:
				s.target(lex, expr)
			case lexer.TokenEqual:
				s.assignment(lex, expr)
			}
			log.Printf("debug: %v\n", expr)
			continue
		default:
			log.Printf("info:Build: %v\n", lex.TokenToStr(t))
		}
	}
}

func (s *Source) Print() {
	for _, node := range s.Tree {
		switch node.(type) {
		case *NodeTarget:
			tg := node.(*NodeTarget)
			log.Printf("info:Print:target %v\n", tg.ID.Value())
			for _, expr := range tg.Prerequisites {
				log.Printf("info:Print:prerequisite %v\n", expr.Value())
			}
			for _, expr := range tg.Recipe {
				log.Printf("info:Print:recipe %v\n", expr.Value())
			}
		default:
			log.Printf("err: unknown node type\n")
		}
	}
}

func (s *Source) skipSpaces(lex lexer.Tokenizer) lexer.Token {
	t := lex.GetToken()
	for t.Type == lexer.TokenSpace {
		t = lex.GetToken()
	}
	return t
}
