package pkg

import (
	"fmt"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/diff"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"gopkg.in/yaml.v3"
)

func TestConstants(t *testing.T) {

	p, err := parse.File("./tmpl/constant2-template.yaml")
	if err != nil {
		t.Fatal(err)
	}

	tmpl, err := Template(p, ".", nil)
	if err != nil {
		t.Fatal(err)
	}

	et, err := parse.File("./tmpl/constant2-expect.yaml")
	if err != nil {
		t.Fatal(err)
	}

	d := diff.New(tmpl, et)
	if d.Mode() != "=" {
		t.Errorf("Output does not match expected: %v", d.Format(true))
	}

}

func TestReplaceConstants(t *testing.T) {
	n := &yaml.Node{Kind: yaml.ScalarNode, Value: "${Rain::Test}"}
	constants := make(map[string]*yaml.Node)
	constants["Test"] = &yaml.Node{Kind: yaml.ScalarNode, Value: "Foo"}
	err := replaceConstants(n, constants)
	if err != nil {
		t.Fatal(err)
	}
	if n.Value != "Foo" {
		t.Fatalf("Expected Foo, got %s", n.Value)
	}
}

func TestIsSubNeeded(t *testing.T) {
	cases := make(map[string]bool)
	cases["ABC"] = false
	cases["${A}bc"] = true
	cases["${Rain::Something}"] = true
	cases[""] = false
	cases["${Abc.Def"] = true
	cases["${!saml:sub}"] = false
	cases["${!Literal}-abc"] = false
	cases["$foo$bar"] = false

	for k, v := range cases {
		if parse.IsSubNeeded(k) != v {
			t.Errorf("IsSubNeeded(%s) should be %v", k, v)
		}
	}
}

func TestUnsupportedConstantInterpolation(t *testing.T) {
	for _, value := range []string{
		`!Sub ["${Env}-${X}", {X: billing}]`,
		`!Ref Env`,
		`{Name: billing}`,
		`[billing, uploads]`,
	} {
		for _, alias := range []string{"Rain", "Const"} {
			for _, inConstant := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/inConstant=%t", value, alias, inConstant), func(t *testing.T) {
					constants := ""
					use := fmt.Sprintf("${%s::Unsupported}", alias)
					if inConstant {
						use = "${Env}"
						constants = fmt.Sprintf("    Derived: !Sub ${%s::Unsupported}-uploads\n", alias)
					}
					input := fmt.Sprintf(`Parameters:
  Env:
    Type: String
Rain:
  Constants:
    Unsupported: %s
%sOutputs:
  Bad:
    Value: !Sub %s-uploads
  Good:
    Value: !Sub ${Env}-ok
`, value, constants, use)
					tmpl, err := parse.String(input)
					if err != nil {
						t.Fatal(err)
					}
					_, err = Template(tmpl, ".", nil)
					if err == nil || !strings.Contains(err.Error(), "constant Unsupported cannot be interpolated as a string") {
						t.Fatalf("expected unsupported constant error, got %v", err)
					}
				})
			}
		}
	}
}
