package skaff

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

const (
	typeResource   = "resource"
	typeDatasource = "data_source"
)

var funcMap = template.FuncMap{
	"camel":      strcase.ToCamel,
	"lowerCamel": strcase.ToLowerCamel,
	"kebab":      strcase.ToKebab,
}

func Generate(t, n string) error {
	if t != typeResource && t != typeDatasource {
		return fmt.Errorf("invalid type. Must be one of: %s, %s", typeResource, typeDatasource)
	}

	if n == "" {
		return errors.New("name cannot be empty")
	}

	if strings.ToLower(strcase.ToSnake(n)) != n {
		return errors.New("name must be in snake_case lowercase")
	}

	data := struct {
		Name string
	}{
		Name: n,
	}

	err := writeTemplate(t+".tmpl", fmt.Sprintf("%s_%s.go", n, t), data)
	if err != nil {
		return err
	}

	return writeTemplate(t+"_test.tmpl", fmt.Sprintf("%s_%s_test.go", n, t), data)
}

func writeTemplate(tmplName, fileName string, data any) error {
	tmpl, err := template.New(tmplName).Funcs(funcMap).ParseFiles("templates/" + tmplName)
	if err != nil {
		return err
	}

	out := filepath.Join("../internal/provider", fileName)
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}

	log.Println("Created", fileName)
	return nil
}
