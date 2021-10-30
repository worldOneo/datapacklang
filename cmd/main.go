package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/worldOneo/datapacklang/ast"
	"github.com/worldOneo/datapacklang/tokens"
	"github.com/worldOneo/datapacklang/translator"
)

func main() {
	var file string
	var overwrite bool
	flag.StringVar(&file, "file", "main.dpl", "Defines the file to translate to mcfunction")
	flag.BoolVar(&overwrite, "overwrite", false, "If overwrite is defined already existing .mcfunction files will be overwritten by the compilation of a .dpl file")

	flag.Parse()

	info, err := os.Stat(file)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		err := TranslateFile(file, overwrite)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := filepath.Walk(file, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if info.IsDir() {
				return nil
			}
			return TranslateFile(path, overwrite)
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(0)
}

func TranslateFile(path string, overwrite bool) error {
	if filepath.Ext(path) != ".dpl" {
		return nil
	}
	newFile := strings.TrimSuffix(path, filepath.Ext(path)) + ".mcfunction"
	if !overwrite {
		info, err := os.Stat(newFile)
		if err == nil {
			return fmt.Errorf("File %s already exists use -overwrite to overwrite the old file", newFile)
		}
		if !os.IsNotExist(err) {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("Path %s is directory but file required", path)
		}
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	code := string(content)
	tokens, err := tokens.Lexer(code)
	if err != nil {
		return err
	}
	parsed, err := ast.Parse(tokens)
	if err != nil {
		return err
	}
	translator := translator.New()
	res, err := translator.Translate(parsed)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(newFile, os.O_TRUNC|os.O_CREATE, 0o660)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(strings.Join(res, "\r\n")))
	return err
}
