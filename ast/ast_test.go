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
			"t1",
			args{lexed: tokens.Lexerp(`
			route("/test/", yeet("me", "out"))
			`)},
			Block{
				Body: []Node{
					Expression{"route", []Node{
						String{"/test/"},
						Expression{"yeet", []Node{
							String{"me"},
							String{"out"}},
						},
					},
					},
				},
			},
			false,
		},
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
					StoreAssign{"test", "store", tokens.OperationSet, Int{100}},
					StoreAssign{"test", "store", tokens.OperationAdd, Int{1}},
					StoreAssign{"test", "store", tokens.OperationAdd, Int{120}},
					StoreAssign{"test", "store", tokens.OperationSub, Int{2}},
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
					StoreAssign{"b", "a", tokens.OperationSet, StoreAccess{"d", "c"}},
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
