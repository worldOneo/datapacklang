package tokens

import (
	"reflect"
	"testing"
)

func TestCodeLexer_Lexer(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    []Token
		wantErr bool
	}{
		{
			"t1",
			`scope {
				print("test")
			}`,
			[]Token{
				identifierToken("scope", 0), scopeOpenToken(0),
				identifierToken("print", 1), {ParenOpen, "(", 0, 0, 1}, stringToken("test", 1), {ParenClosed, ")", 0, 0, 1},
				scopeClosedToken(2),
			},
			false,
		},
		{
			"operations",
			`store[test] = 1
			store[test] += 1
			store[test]++
			`,
			[]Token{
				identifierToken("store", 0), {IndexOpen, "[", 0, 0, 0}, identifierToken("test", 0), {IndexClosed, "]", 0, 0, 0}, {OperationAssignment, "=", OperationSet, 0, 0}, {Integer, "1", 1, 0, 0},
				identifierToken("store", 1), {IndexOpen, "[", 0, 0, 1}, identifierToken("test", 1), {IndexClosed, "]", 0, 0, 1}, {OperationAssignment, "+=", OperationAdd, 0, 1}, {Integer, "1", 1, 0, 1},
				identifierToken("store", 2), {IndexOpen, "[", 0, 0, 2}, identifierToken("test", 2), {IndexClosed, "]", 0, 0, 2}, {OperationAssignment, "++", OperationInc, 0, 2},
			},
			false,
		},
		{
			"calculations",
			`a[b] = c[d]+1`,
			[]Token{
				identifierToken("a", 0), {IndexOpen, "[", 0, 0, 0}, identifierToken("b", 0), {IndexClosed, "]", 0, 0, 0}, {OperationAssignment, "=", OperationSet, 0, 0},
				identifierToken("c", 0), {IndexOpen, "[", 0, 0, 0}, identifierToken("d", 0), {IndexClosed, "]", 0, 0, 0},
				{Operation, "+", OperationAdd, 0, 0}, {Integer, "1", 1, 0, 0},
			},
			false,
		},
		{
			"calculations primitives",
			`a[b] = 1+2`,
			[]Token{
				identifierToken("a", 0), {IndexOpen, "[", 0, 0, 0}, identifierToken("b", 0), {IndexClosed, "]", 0, 0, 0},
				{OperationAssignment, "=", OperationSet, 0, 0}, {Integer, "1", 1, 0, 0}, {Operation, "+", OperationAdd, 0, 0}, {Integer, "2", 2, 0, 0},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Lexer(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("WordParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WordParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
