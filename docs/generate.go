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
	return s
}

func main() {
	console.NoColour = true

	// No "Auto generated on <date>" footer, so regenerating on another day
	// does not change every page and the CI freshness check stays stable
	rain.Cmd.DisableAutoGenTag = true

	err := doc.GenMarkdownTreeCustom(rain.Cmd, "./", emptyStr, identity)
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

	// Users link cfnrain as rain to keep existing scripts working, so complete both names
	alsoComplete("bash_completion.sh", "-F __start_cfnrain cfnrain\n", "-F __start_cfnrain cfnrain rain\n")
	alsoComplete("zsh_completion.sh", "#compdef cfnrain\n", "#compdef cfnrain rain\n")
	alsoComplete("zsh_completion.sh", "compdef _cfnrain cfnrain\n", "compdef _cfnrain cfnrain rain\n")
}

// alsoComplete rewrites a generated completion script and panics when
// the expected line is missing, so a cobra change cannot silently drop
// completion for the rain name
func alsoComplete(path, old, new string) {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	s := string(b)
	if !strings.Contains(s, old) {
		panic("completion line not found in " + path + ": " + old)
	}
	err = os.WriteFile(path, []byte(strings.ReplaceAll(s, old, new)), 0644)
	if err != nil {
		panic(err)
	}
}
