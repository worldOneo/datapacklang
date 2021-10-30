package ast

import (
	"fmt"

	"github.com/worldOneo/datapacklang/tokens"
)

type Type uint64

const (
	TypeCall Type = iota
	TypeVariable
)

type Node interface{}

type Variable struct {
	Name string
}

type String struct {
	Value string
}

type Float struct {
	Value float64
}

type Int struct {
	Value int
}

type Block struct {
	Body []Node
}

type CreateStore struct {
	Identifier string
}

type Calculation struct {
	First    Node
	Operator tokens.OperationType
	Second   Node
}

type StoreAssign struct {
	Identifier string
	Store      string
	Operation  tokens.OperationType
	Value      Node
}

type StoreAccess struct {
	Identifier string
	Store      string
}

type Expression struct {
	Identifier string
	ArgList    []Node
}

type Program = Block

type Parser struct {
	tokens []tokens.Token
	index  int
}

func Parse(lexed []tokens.Token) (Node, error) {
	parser := Parser{
		tokens: lexed,
	}
	return parser.parse()
}

func (P *Parser) parse() (Node, error) {
	l := 64
	body := make([]Node, l)
	bindex := 0
	peek, peeked := P.peek()
	returnOnScopeClose := peeked && peek.Type == tokens.ScopeOpen

	if returnOnScopeClose {
		P.next()
	}

	for P.index < len(P.tokens) {
		if returnOnScopeClose {
			peek, peeked = P.peek()
			if peeked && peek.Type == tokens.ScopeClosed {
				P.next()
				break
			}
		}
		node, err := P.pullValue()
		if err != nil {
			return nil, err
		}
		body[bindex] = node
		bindex++
		if bindex >= l {
			old := body
			l *= 2
			body = make([]Node, l)
			copy(body, old)
		}
	}
	return Block{body[0:bindex]}, nil
}

func (P *Parser) peek() (tokens.Token, bool) {
	if P.index < len(P.tokens) {
		return P.tokens[P.index], true
	}
	return tokens.Token{}, false
}

func (P *Parser) next() (tokens.Token, bool) {
	if P.index < len(P.tokens) {
		P.index++
		return P.tokens[P.index-1], true
	}
	return tokens.Token{}, false
}

func (P *Parser) argList() ([]Node, error) {
	args := make([]Node, 0)
	requiresComma := false

	for peek, peeked := P.peek(); peeked && peek.Type != tokens.ParenClosed; peek, peeked = P.peek() {
		if requiresComma && peek.Type == tokens.Comma {
			requiresComma = false
			P.next()
			continue
		} else if requiresComma || peek.Type == tokens.Comma {
			return nil, fmt.Errorf("unexpected comma")
		}
		arg, err := P.pullValue()
		requiresComma = true
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	P.next()
	return args, nil
}

func (P *Parser) pullValue() (Node, error) {
	next, has := P.next()
	if !has {
		return nil, fmt.Errorf("Expected value")
	}

	peek, _ := P.peek()
	switch next.Type {
	case tokens.Identifier:
		if peek.Type == tokens.ParenOpen {
			P.next()
			args, err := P.argList()
			if err != nil {
				return nil, err
			}
			return Expression{next.Content, args}, nil
		} else if peek.Type == tokens.IndexOpen {
			P.next()
			identifier, ok := P.next()
			if !ok || identifier.Type != tokens.Identifier {
				break
			}
			closedIndex, ok := P.next()
			if !ok || closedIndex.Type != tokens.IndexClosed {
				break
			}
			operation, ok := P.peek()
			if !ok || operation.Type != tokens.OperationAssignment {
				if ok && operation.Type == tokens.Operation {
					P.next()
					value, err := P.pullValue()
					if err != nil {
						return nil, err
					}
					return Calculation{StoreAccess{identifier.Content, next.Content}, operation.ValueInt, value}, nil
				}
				return StoreAccess{identifier.Content, next.Content}, nil
			}
			P.next()
			if operation.ValueInt == tokens.OperationInc {
				return StoreAssign{identifier.Content, next.Content, tokens.OperationAdd, Int{1}}, nil
			}

			if operation.ValueInt == tokens.OperationDec {
				return StoreAssign{identifier.Content, next.Content, tokens.OperationSub, Int{1}}, nil
			}
			value, err := P.pullValue()
			if err != nil {
				return nil, err
			}
			return StoreAssign{identifier.Content, next.Content, operation.ValueInt, value}, nil
		}
	case tokens.Create:
		if peek.Type != tokens.Identifier {
			break
		}
		P.next()
		name, ok := P.next()
		if !ok || name.Type != tokens.Identifier {
			break
		}
		if peek.Content == "store" {
			return CreateStore{name.Content}, nil
		}
	case tokens.Float:
		return Float{next.ValueFloat}, nil
	case tokens.Integer:
		peek, ok := P.peek()
		if ok && peek.Type == tokens.Operation {
			P.next()
			value, err := P.pullValue()
			if err != nil {
				return nil, err
			}
			return Calculation{Int{next.ValueInt}, peek.ValueInt, value}, nil
		}
		return Int{next.ValueInt}, nil
	case tokens.String:
		return String{next.Content}, nil
	}
	return nil, fmt.Errorf("Identifier Expected line: %d at '%s'", next.Line, next.Content)
}
