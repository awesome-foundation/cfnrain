package format_test

import (
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
)

// A template with CRLF line endings must format the same as with LF.
// Based on the test in aws-cloudformation/rain#782 by @cv-dote.
func TestCRLFComments(t *testing.T) {
	cases := map[string]string{
		"comments only": "# Comment line 1\n" +
			"# Comment line 2\n" +
			"Description: Hello\n",
		"blank lines around comments": "AWSTemplateFormatVersion: 2010-09-09\n" +
			"\n" +
			"# Comment line 1\n" +
			"# Comment line 2\n" +
			"# Comment line 3\n" +
			"\n" +
			"Description: Hello World\n",
		"comments in mappings and scalars": "# Head\n" +
			"Resources:\n" +
			"  # Bucket comment 1\n" +
			"  # Bucket comment 2\n" +
			"  Bucket:\n" +
			"    Type: AWS::S3::Bucket # line comment\n" +
			"    Properties:\n" +
			"      BucketName: |\n" +
			"        line one\n" +
			"        line two\n" +
			"      Tags:\n" +
			"        - Key: \"quoted\n" +
			"            folded\"\n" +
			"          Value: >\n" +
			"            folded\n" +
			"            text\n" +
			"# Foot\n",
	}

	for name, lf := range cases {
		t.Run(name, func(t *testing.T) {
			crlf := strings.ReplaceAll(lf, "\n", "\r\n")

			tmplLF, err := parse.String(lf)
			if err != nil {
				t.Fatalf("parse LF: %v", err)
			}
			tmplCRLF, err := parse.String(crlf)
			if err != nil {
				t.Fatalf("parse CRLF: %v", err)
			}

			outLF := format.String(tmplLF, format.Options{})
			outCRLF := format.String(tmplCRLF, format.Options{})

			for _, line := range strings.Split(lf, "\n") {
				if i := strings.Index(line, "#"); i >= 0 {
					comment := line[i:]
					if !strings.Contains(outCRLF, comment) {
						t.Errorf("CRLF output lost comment %q:\n%s", comment, outCRLF)
					}
				}
			}

			if outCRLF != outLF {
				t.Errorf("CRLF and LF outputs differ.\nCRLF:\n%q\nLF:\n%q", outCRLF, outLF)
			}
		})
	}
}

// Mixed line endings must format like LF and keep the number of
// line breaks in values. A lone CR is a line break in YAML.
func TestMixedLineEndings(t *testing.T) {
	cases := map[string]string{
		"Description: |\n  a\r\r\n  b\n": "Description: |\n  a\n\n  b\n",
		"Description: >\n  a\r\r\n  b\n": "Description: >\n  a\n\n  b\n",
		"Description: \"a\r\r\n  b\"\n":  "Description: \"a\n\n  b\"\n",
		"Description: a\r\r\n  b\n":      "Description: a\n\n  b\n",
		"# c1\r# c2\rDescription: a\r":   "# c1\n# c2\nDescription: a\n",
		"# c1\r\n# c2\nDescription: a\r": "# c1\n# c2\nDescription: a\n",
	}

	for input, lf := range cases {
		tmpl, err := parse.String(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		tmplLF, err := parse.String(lf)
		if err != nil {
			t.Fatalf("parse %q: %v", lf, err)
		}
		got := format.String(tmpl, format.Options{})
		want := format.String(tmplLF, format.Options{})
		if got != want {
			t.Errorf("input %q:\ngot  %q\nwant %q", input, got, want)
		}
	}
}

// Line ending normalization must not touch UTF-16 input, where a 0x0d
// byte can be half of another character.
func TestUTF16NotNormalized(t *testing.T) {
	// UTF-16LE with BOM: "Description: č\n" (č is U+010D, bytes 0d 01)
	input := "\xff\xfe" +
		"D\x00e\x00s\x00c\x00r\x00i\x00p\x00t\x00i\x00o\x00n\x00:\x00 \x00" +
		"\x0d\x01\n\x00"
	tmpl, err := parse.String(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := format.String(tmpl, format.Options{})
	if want := "Description: č\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
