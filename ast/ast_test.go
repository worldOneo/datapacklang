package ast

import (
	"reflect"
	"testing"

	"github.com/worldOneo/datapacklang/tokens"
)

func TestParse(t *testing.T) {
	type args struct {
		lexed []tokens.Token
	}
	tests := []struct {
		name    string
		args    args
		want    Node
		wantErr bool
	}{
		{
			"value assignment",
			args{lexed: tokens.Lexerp(`
			store[test] = 100
			store[test]++
			store[test] += 120
			store[test] -= 2
			`)},
			Block{
				Body: []Node{
					MakeStoreAssign("store", "test", true, tokens.OperationSet, Int{100}),
					MakeStoreAssign("store", "test", true, tokens.OperationAdd, Int{1}),
					MakeStoreAssign("store", "test", true, tokens.OperationAdd, Int{120}),
					MakeStoreAssign("store", "test", true, tokens.OperationSub, Int{2}),
				},
			},
			false,
		},
		{
			"store assignment",
			args{lexed: tokens.Lexerp(`
				a[b] = c[d]
			`)},
			Block{
				Body: []Node{
					MakeStoreAssign("a", "b", true, tokens.OperationSet, MakeStoreAccess("c", "d", true)),
				},
			},
			false,
		},
		{
			"calculations",
			args{tokens.Lexerp(`a[b] = c[d] + 3`)},
			Block{
				[]Node{
					MakeStoreAssign("a", "b", true, tokens.OperationSet, Calculation{MakeStoreAccess("c", "d", true), tokens.OperationAdd, Int{3}}),
				},
			},
			false,
		},
		{
			"calculations primitives",
			args{tokens.Lexerp(`a[b] = 1+2`)},
			Block{
				[]Node{
					MakeStoreAssign("a", "b", true, tokens.OperationSet, Calculation{Int{1}, tokens.OperationAdd, Int{2}}),
				},
			},
			false,
		},
		{
			"if",
			args{tokens.Lexerp("if 1 < 2 { 'say hi' }")},
			Block{
				[]Node{
					If{Int{1}, tokens.OperationLt, Int{2}, false, Block{
						[]Node{
							String{"say hi"},
						},
					}},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.lexed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
