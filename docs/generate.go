package main

//go:generate go run generate.go

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra/doc"

	"github.com/awesome-foundation/cfnrain/internal/cmd/rain"
	"github.com/awesome-foundation/cfnrain/internal/console"
)

var tmpl *template.Template

func init() {
	var err error

	tmpl = template.New("README.tmpl")

	tmpl = tmpl.Funcs(template.FuncMap{
		"pad": func(s string, n int) string {
			return strings.Repeat(" ", n-len(s))
		},
	})
	if err != nil {
		panic(err)
	}

	tmpl, err = tmpl.ParseFiles("README.tmpl")
	if err != nil {
		panic(err)
	}
}

func emptyStr(s string) string {
	return ""
}

func identity(s string) string {
	if s == "cfnrain.md" {
		return "index.md"
	}

	return s
}

func main() {
	console.NoColour = true

	err := doc.GenMarkdownTreeCustom(rain.Cmd, "./", emptyStr, identity)
	if err != nil {
		panic(err)
	}

	err = os.Rename("cfnrain.md", "index.md")
	if err != nil {
		panic(err)
	}

	// Generate usage
	usage := bytes.Buffer{}
	rain.Cmd.SetOut(&usage)
	rain.Cmd.Usage()

	// Generate README
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, map[string]string{
		"Usage": usage.String(),
	})
	if err != nil {
		panic(err)
	}

	os.WriteFile("../README.md", buf.Bytes(), 0644)

	rain.Cmd.GenBashCompletionFile("bash_completion.sh")
	rain.Cmd.GenZshCompletionFile("zsh_completion.sh")
}
