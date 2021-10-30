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

type As struct {
	Selector string
	Body     Block
}

type If struct {
	First      Node
	Comparator tokens.OperationType
	Second     Node
	Not        bool
	Body       Block
}

type Index struct {
	Identifier string
	IsVar      bool
}

type StoreAssign struct {
	Identifier Index
	Store      string
	Operation  tokens.OperationType
	Value      Node
}

type StoreAccess struct {
	Identifier Index
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
			if !ok || (identifier.Type != tokens.Identifier && identifier.Type != tokens.String) {
				break
			}
			isVar := identifier.Type == tokens.Identifier
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
					return Calculation{StoreAccess{Index{identifier.Content, isVar}, next.Content}, operation.ValueInt, value}, nil
				}
				return StoreAccess{Index{identifier.Content, isVar}, next.Content}, nil
			}
			P.next()
			if operation.ValueInt == tokens.OperationInc {
				return StoreAssign{Index{identifier.Content, isVar}, next.Content, tokens.OperationAdd, Int{1}}, nil
			}

			if operation.ValueInt == tokens.OperationDec {
				return StoreAssign{Index{identifier.Content, isVar}, next.Content, tokens.OperationSub, Int{1}}, nil
			}
			value, err := P.pullValue()
			if err != nil {
				return nil, err
			}
			return StoreAssign{Index{identifier.Content, isVar}, next.Content, operation.ValueInt, value}, nil
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
	case tokens.If:
		not := false
		if peek.Type == tokens.Not {
			not = true
			P.next()
		}
		first, err := P.pullValue()
		if err != nil {
			return nil, err
		}
		comparator, ok := P.next()
		if !ok || comparator.Type != tokens.OperationComp {
			return nil, fmt.Errorf("Comparator expected")
		}
		second, err := P.pullValue()
		if err != nil {
			return nil, err
		}
		body, err := P.parse()
		if err != nil {
			return nil, err
		}
		if _, ok := body.(Block); !ok {
			return nil, fmt.Errorf("If requires body line: %d", next.Line)
		}
		return If{first, comparator.ValueInt, second, not, body.(Block)}, nil
	case tokens.As:
		if peek.Type != tokens.String {
			return nil, fmt.Errorf("As requires selector line: %d", next.Line)
		}
		P.next()
		body, err := P.parse()
		if err != nil {
			return nil, err
		}
		if _, ok := body.(Block); !ok {
			return nil, fmt.Errorf("As requires body line: %d", next.Line)
		}
		return As{peek.Content, body.(Block)}, nil
	}
	return nil, fmt.Errorf("Identifier Expected line: %d at '%s'", next.Line, next.Content)
}

func MakeStoreAccess(store, identifier string, isVar bool) StoreAccess {
	return StoreAccess{Identifier: Index{Identifier: identifier, IsVar: isVar}, Store: store}
}

func MakeStoreAssign(store, identifier string, isVar bool, operation tokens.OperationType, value Node) StoreAssign {
	return StoreAssign{Identifier: Index{Identifier: identifier, IsVar: isVar}, Store: store, Operation: operation, Value: value}
}
