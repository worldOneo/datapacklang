package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/worldOneo/datapacklang/ast"
	"github.com/worldOneo/datapacklang/tokens"
	"github.com/worldOneo/datapacklang/translator"
)

func main() {
	var file string
	flag.StringVar(&file, "file", "main.dpl", "Defines the file to translate to mcfunction")
	flag.Parse()
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	code := string(content)
	tokens, err := tokens.Lexer(code)
	if err != nil {
		log.Fatal(err)
	}
	parsed, err := ast.Parse(tokens)
	if err != nil {
		log.Fatal(err)
	}
	translator := translator.New()
	res, err := translator.Translate(parsed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(res, "\r\n"))
}
