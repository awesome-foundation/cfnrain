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
	err := replaceConstants(n, constants, nil, false)
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

func TestConstantsPreserveUnclosedVariables(t *testing.T) {
	for _, value := range []string{
		`'^\${[A-Z]+'`,
		`{Script: 'echo ${'}`,
	} {
		t.Run(value, func(t *testing.T) {
			input := "Rain:\n  Constants:\n    Literal: " + value + "\nOutputs:\n  Literal:\n    Value: !Rain::Constant Literal\n"
			expected := "Outputs:\n  Literal:\n    Value: " + value + "\n"
			assertPackagedConstants(t, input, expected)
		})
	}
	t.Run("Sub without constants", func(t *testing.T) {
		input := "Outputs:\n  Literal:\n    Value: !Sub '${AWS::Region} costs ${'\n"
		assertPackagedConstants(t, input, input)
	})
}

func assertPackagedConstants(t *testing.T, input, expected string) {
	t.Helper()
	tmpl, err := parse.String(input)
	if err != nil {
		t.Fatal(err)
	}
	packaged, err := Template(tmpl, ".", nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := parse.String(expected)
	if err != nil {
		t.Fatal(err)
	}
	d := diff.New(packaged, want)
	if d.Mode() != "=" {
		t.Errorf("Output does not match expected: %v", d.Format(true))
	}
}

func TestMissingConstantInterpolation(t *testing.T) {
	tmpl, err := parse.String("Outputs:\n  Missing:\n    Value: !Sub '${Rain::Missing}'\n")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Template(tmpl, ".", nil)
	if err == nil || !strings.Contains(err.Error(), "did not find Constant Missing") {
		t.Fatalf("expected missing constant error, got %v", err)
	}
}

func TestNestedSubConstantPreservesEscapes(t *testing.T) {
	input := `Rain:
  Constants:
    Escaped: !Sub '${!Literal}'
Outputs:
  Escaped:
    Value: !Sub '${Rain::Escaped}'
  Suffixed:
    Value: !Sub '${Const::Escaped}-uploads'
`
	expected := `Outputs:
  Escaped:
    Value: !Sub '${!Literal}'
  Suffixed:
    Value: !Sub '${!Literal}-uploads'
`
	assertPackagedConstants(t, input, expected)
}

func TestPlainConstantRejectsSubInterpolation(t *testing.T) {
	for _, alias := range []string{"Rain", "Const"} {
		for _, value := range []string{`"${%s::B}"`, `{Name: "${%s::B}"}`, `["${%s::B}"]`} {
			t.Run(alias+"/"+value, func(t *testing.T) {
				input := fmt.Sprintf(`Parameters:
  Env:
    Type: String
Rain:
  Constants:
    B: !Sub ${Env}-b
    C: %s
Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Rain::Constant C
`, fmt.Sprintf(value, alias))
				tmpl, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				_, err = Template(tmpl, ".", nil)
				want := "constant C: constant B uses Fn::Sub and cannot be interpolated into a plain string"
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("expected %q, got %v", want, err)
				}
			})
		}
	}
}

func TestConstantInterpolationPreservesValidValues(t *testing.T) {
	input := `Parameters:
  Env:
    Type: String
Rain:
  Constants:
    Plain: billing
    Derived: "${Rain::Plain}-uploads"
    B: !Sub ${Env}-b
    C: !Sub ${Const::B}-c
    Object:
      Name: !Sub ${Rain::B}-object
Outputs:
  Plain:
    Value: !Rain::Constant Derived
  Sub:
    Value: !Rain::Constant C
  Object:
    Value: !Rain::Constant Object
`
	expected := `Parameters:
  Env:
    Type: String
Outputs:
  Plain:
    Value: billing-uploads
  Sub:
    Value: !Sub ${Env}-b-c
  Object:
    Value:
      Name: !Sub ${Env}-b-object
`
	assertPackagedConstants(t, input, expected)
}

func TestSubConstantReviewRegressions(t *testing.T) {
	for _, value := range []string{`"${Rain::B}"`, `{"${Rain::B}": unused}`, `"${Rain::B}"` + "\n    D: '${Const::C}'"} {
		t.Run("unused/"+value, func(t *testing.T) {
			input := "Rain:\n  Constants:\n    B: !Sub '${Env}-b'\n    C: " + value + "\nOutputs:\n  Value:\n    Value: !Rain::Constant B\n"
			assertPackagedConstants(t, input, "Outputs:\n  Value:\n    Value: !Sub '${Env}-b'\n")
		})
	}
	t.Run("sequence Sub", func(t *testing.T) {
		input := "Rain:\n  Constants:\n    B: !Sub '${Env}-b'\n    C: !Sub ['${Rain::B}-${X}', {X: !Ref Env}]\nOutputs:\n  Value:\n    Value: !Rain::Constant C\n"
		assertPackagedConstants(t, input, "Outputs:\n  Value:\n    Value: !Sub ['${Env}-b-${X}', {X: !Ref Env}]\n")
	})
	for _, value := range []string{`!Sub literal`, `!Sub '${Const::A}-x'`} {
		t.Run("resolved/"+value, func(t *testing.T) {
			input := "Rain:\n  Constants:\n    A: a\n    B: " + value + "\n    C: '${Const::B}-y'\nOutputs:\n  Value:\n    Value: !Rain::Constant C\n"
			want := "literal-y"
			if strings.Contains(value, "Const::A") {
				want = "a-x-y"
			}
			assertPackagedConstants(t, input, "Outputs:\n  Value:\n    Value: "+want+"\n")
		})
	}
}

func TestInvalidPlainConstantUse(t *testing.T) {
	for _, use := range []string{`!Rain::Constant C`, `!Ref Const::C`, `!Sub '${Rain::C}'`, `!Sub '${Const::C}'`, `!Rain::Constant D`} {
		for _, derived := range []string{`'${Const::C}'`, `!Rain::Constant C`, `!Ref Const::C`} {
			t.Run(use+"/"+derived, func(t *testing.T) {
				input := "Rain:\n  Constants:\n    B: !Sub '${Env}-b'\n    C: '${Rain::B}'\n    D: " + derived + "\nOutputs:\n  Value:\n    Value: " + use + "\n"
				tmpl, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				_, err = Template(tmpl, ".", nil)
				if err == nil || !strings.Contains(err.Error(), "constant C: constant B uses Fn::Sub") || !strings.Contains(err.Error(), "Use !Sub") {
					t.Fatalf("expected constant error with remedy, got %v", err)
				}
			})
		}
	}
}

func TestSubConstantStillNeedsSub(t *testing.T) {
	for _, value := range []string{`!Sub '${!Literal}'`, `!Sub ['${X}', {X: '${Rain::B}'}]`} {
		t.Run(value, func(t *testing.T) {
			input := "Rain:\n  Constants:\n    B: !Sub '${Env}-b'\n    C: " + value + "\n    D: '${Rain::C}'\nOutputs:\n  Value:\n    Value: !Rain::Constant "
			if strings.Contains(value, "${!Literal}") {
				input += "D\n"
			} else {
				input += "C\n"
			}
			tmpl, err := parse.String(input)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Template(tmpl, ".", nil)
			if err == nil || !strings.Contains(err.Error(), "cannot be interpolated into a plain string") {
				t.Fatalf("expected plain string error, got %v", err)
			}
		})
	}
}

func TestDirectLiteralSubConstant(t *testing.T) {
	assertPackagedConstants(t,
		"Rain:\n  Constants:\n    B: !Sub literal\nOutputs:\n  Value:\n    Value: !Rain::Constant B\n",
		"Outputs:\n  Value:\n    Value: literal\n")
}

func TestDeferredConstantErrorDoesNotHideMissingConstant(t *testing.T) {
	for _, use := range []string{"ok", "!Rain::Constant C"} {
		for _, first := range []string{`"${Rain::B}"`, "!Rain::Constant Deferred", "!Ref Const::Deferred"} {
			t.Run(use+"/"+first, func(t *testing.T) {
				input := `Rain:
  Constants:
    B: !Sub ${Env}-b
    Deferred: "${Rain::B}"
    C: [` + first + `, !Sub "${Rain::Missing}"]
Outputs:
  Value:
    Value: ` + use + "\n"
				tmpl, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				_, err = Template(tmpl, ".", nil)
				want := "constant C: did not find Constant Missing"
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("expected %q, got %v", want, err)
				}
			})
		}
	}
}

func TestDeferredConstantErrorDoesNotHideMissingWord(t *testing.T) {
	for _, value := range []string{
		`"${Rain::B}-${Rain::Missing}"`,
		`!Sub "${Rain::Deferred}-${Rain::Missing}"`,
	} {
		for _, use := range []string{"ok", "!Rain::Constant C"} {
			t.Run(value+"/"+use, func(t *testing.T) {
				input := `Rain:
  Constants:
    B: !Sub ${Env}-b
    Deferred: "${Rain::B}"
    C: ` + value + `
Outputs:
  Value:
    Value: ` + use + "\n"
				tmpl, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				_, err = Template(tmpl, ".", nil)
				want := "constant C: did not find Constant Missing"
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("expected %q, got %v", want, err)
				}
			})
		}
	}
	t.Run("template Sub", func(t *testing.T) {
		tmpl, err := parse.String(`Rain:
  Constants:
    B: !Sub ${Env}-b
    Deferred: "${Rain::B}"
Outputs:
  Value:
    Value: !Sub "${Rain::Deferred}-${Rain::Missing}"
`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = Template(tmpl, ".", nil)
		want := "did not find Constant Missing"
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("expected %q, got %v", want, err)
		}
	})
}
