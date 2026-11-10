package format_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
)

func TestStrictBooleanAliasOutput(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	for _, value := range []string{"yes", "no", "true", "false", "1", "first\nsecond", "\n\nfirst\nsecond", "first\nsecond\n\n", "\n\nfirst\nsecond\n\n"} {
		t.Run(strconv.Quote(value), func(t *testing.T) {
			input := "Metadata:\n  V: " + strconv.Quote(value) + "\n"
			var want string
			for _, style := range []string{"strict-booleans", "strict-boolean"} {
				template, err := parse.String(input)
				if err != nil {
					t.Fatal(err)
				}
				format.NodeStyle = style
				output := format.String(template, format.Options{})
				if style == "strict-booleans" {
					want = output
				} else if output != want {
					t.Fatalf("alias output:\n%s\nwant:\n%s", output, want)
				}
			}
		})
	}
}

func TestInvalidNodeStyleListsNames(t *testing.T) {
	oldStyle := format.NodeStyle
	t.Cleanup(func() { format.NodeStyle = oldStyle })
	template, err := parse.String("Metadata: {}\n")
	if err != nil {
		t.Fatal(err)
	}
	format.NodeStyle = "invalid"
	defer func() {
		message := fmt.Sprint(recover())
		for _, name := range []string{"invalid --node-style: invalid", "tagged", "doublequoted", "singlequoted", "literal", "folded", "strict-boolean", "strict-booleans", "quotescalars", "original", "flow"} {
			if !strings.Contains(message, name) {
				t.Errorf("error %q does not contain %q", message, name)
			}
		}
	}()
	format.String(template, format.Options{})
	t.Fatal("invalid style did not panic")
}
