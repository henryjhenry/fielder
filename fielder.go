package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

var (
	typeFlag = flag.String("type", "", "special type, required.")
	tagFlag  = flag.String("tag", "", "special struct tag, if not specified, the first tag used.")
	src      = flag.String("src", "", "special source file, required.")
	dryrun   = flag.Bool("dryrun", false, "dryrun, if set, result will output to stdout.")
)

type Template struct {
	Package string
	Type    string
	Fields  []Field
}

type Field struct {
	Name         string
	SQLFieldName string
}

func main() {
	flag.Parse()
	var err error
	fp := *src
	if fp == "" || *typeFlag == "" {
		flag.Usage()
		return
	}
	if fp[0] != '/' {
		fp, err = filepath.Abs(fp)
		if err != nil {
			panic(err)
		}
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, fp, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}
	g := gen{
		typeName: *typeFlag,
		tag:      *tagFlag,
		tmpl:     TMPL,
		dryrun:   *dryrun,
		srcFile:  fp,
	}
	// ast.Print(fs, f)
	ast.Walk(&g, f)
	if len(g.fields) == 0 {
		panic("no fields to generate")
	}
	if err := g.generate(); err != nil {
		panic(err)
	}
}

type gen struct {
	pkg      string
	typeName string
	tag      string
	dryrun   bool
	srcFile  string // source file abs path
	tmpl     string

	fields []Field
}

func (g *gen) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		g.pkg = n.Name.Name
	case *ast.TypeSpec:
		if n.Name.Name == g.typeName {
			t, ok := n.Type.(*ast.StructType)
			if !ok {
				panic(fmt.Errorf("type %s is not a struct type", n.Name.Name))
			}
			fields := parse(t)
			g.addFields(fields...)
			return nil
		}
	}
	return g
}

func (g *gen) addFields(fields ...Field) {
	if len(fields) > 0 {
		g.fields = append(g.fields, fields...)
	}
}

func (g *gen) generate() error {
	t := Template{
		Package: g.pkg,
		Type:    g.typeName,
		Fields:  g.fields,
	}
	tmpl, err := template.New("fielder").Parse(g.tmpl)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(buf, t); err != nil {
		return err
	}
	code, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	if g.dryrun {
		_, err := os.Stdout.Write(code)
		if err != nil {
			return err
		}
		return nil
	}
	dir, file := filepath.Split(g.srcFile)
	filename := strings.Split(file, ".")[0]
	target := dir + "/" + filename + "_fields.go"
	if err = os.WriteFile(target, code, 0o666); err != nil {
		return err
	}
	return nil
}

func parse(t *ast.StructType) []Field {
	tp := newTagParser(*tagFlag)
	var fields []Field
	for _, f := range t.Fields.List {
		name := f.Names[0].Name
		sqlFiledName := tp.Parse(f.Tag.Value)
		if !validateFieldName(sqlFiledName) {
			panic(fmt.Errorf("field '%s' tag value '%s' is invalid", name, sqlFiledName))
		}
		if sqlFiledName != "" {
			fields = append(fields, Field{Name: name, SQLFieldName: sqlFiledName})
		}
	}
	return fields
}

func validateFieldName(field string) bool {
	for _, r := range field {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func newTagParser(typ string) TagParser {
	switch typ {
	case "gorm":
		return &gormTagParser{}
	}
	return &defaultTagParser{}
}

type TagParser interface {
	// Parse tag, return sql field name
	Parse(tag string) string
}
