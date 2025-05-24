// ast
package ast

import (
	"fmt"
	"log"
	"os"
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
	Expanded      bool
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
			continue
		default:
			log.Printf("info:Build: %v\n", lex.TokenToStr(t))
		}
	}
}

func (s *Source) expandTarget(tg *NodeTarget) {
	var prerequisites []string
	for _, expr := range tg.Prerequisites {
		switch expr.(type) {
		case *ExprID:
			exprID := expr.(*ExprID)
			prerequisites = append(prerequisites, exprID.ID)
		case *ExprVar:
			// not implemented
		default:
			log.Fatalf("err:expandTarget: unknown expression type\n")
		}
	}
	for _, expr := range tg.Recipe {
		// recipe consists of cmd
		switch expr.(type) {
		case *ExprCmd:
			exprCmd := expr.(*ExprCmd)
			for _, term := range exprCmd.Terms {
				switch term.(type) {
				case *ExprVar:
					exprVar := term.(*ExprVar)
					if exprVar.ExprID.ID == ExprAllPrerequisites {
						// replace with list of prerequisites
						exprVar.Val = strings.Join(prerequisites, "")
					}
				case *ExprID:
				default:
					log.Fatalf("info:Printf: unknown term type\n")
				}
			}
		default:
			log.Fatalf("info:Print: unknown cmd type\n")
		}
	}

}

func (s *Source) Expand() {
	for _, node := range s.Tree {
		switch node.(type) {
		case *NodeTarget:
			tg := node.(*NodeTarget)
			s.expandTarget(tg)
		default:
			log.Printf("err: unknown node type\n")
		}
	}
}

func (s *Source) Print(out *os.File) {
	for _, node := range s.Tree {
		switch node.(type) {
		case *NodeTarget:
			tg := node.(*NodeTarget)
			sTarget := fmt.Sprintf("%s:", tg.ID.Value())
			out.WriteString(sTarget)
			for _, expr := range tg.Prerequisites {
				sPrerequisite := fmt.Sprintf(" %s", expr.Value())
				out.WriteString(sPrerequisite)
			}
			out.WriteString("\n")
			for _, expr := range tg.Recipe {
				sRecipe := fmt.Sprintf("\t%s\n", expr.Value())
				out.WriteString(sRecipe)
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
