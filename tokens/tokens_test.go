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
				identifierToken("print", 1), {ParenOpen, "(", 0,0,1}, stringToken("test", 1), {ParenClosed, ")", 0,0,1},
				scopeClosedToken(2),
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
