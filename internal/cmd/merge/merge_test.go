package merge

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-foundation/cfnrain/cft/diff"
	"github.com/awesome-foundation/cfnrain/cft/format"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/internal/config"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

// merge must not add quotes to a value that the source does not quote
func TestMergeKeepsUnquotedAmbiguousScalars(t *testing.T) {
	dst, err := parse.String("Parameters:\n  Mode:\n    Default: ON\n")
	if err != nil {
		t.Fatal(err)
	}
	src, err := parse.String("Resources: {}\n")
	if err != nil {
		t.Fatal(err)
	}

	merged, err := mergeTemplates(dst, src)
	if err != nil {
		t.Fatal(err)
	}

	actual := format.String(merged, format.Options{})
	if !strings.Contains(actual, "Default: ON\n") {
		t.Errorf("expected Default: ON without quotes, got:\n%s", actual)
	}
}

// Other node styles keep the merge output they had before
func TestMergeOriginalStyleKeepsQuotes(t *testing.T) {
	defer func(s string) { format.NodeStyle = s }(format.NodeStyle)
	format.NodeStyle = "original"

	dst, err := parse.String("Outputs:\n  Flag:\n    Value:\n      Fn::Sub: \"ON\"\n")
	if err != nil {
		t.Fatal(err)
	}
	src, err := parse.String("Resources: {}\n")
	if err != nil {
		t.Fatal(err)
	}

	merged, err := mergeTemplates(dst, src)
	if err != nil {
		t.Fatal(err)
	}

	actual := format.String(merged, format.Options{})
	if !strings.Contains(actual, "Value: !Sub \"ON\"\n") {
		t.Errorf("expected Value: !Sub \"ON\", got:\n%s", actual)
	}
}

func TestMergeTemplatesSuccess(t *testing.T) {
	dst, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "overwritten",
		"Description":              "Line 1",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterGroups": []interface{}{
					map[string]interface{}{
						"Label": map[string]interface{}{
							"default": "Network Configuration",
						},
						"Parameters": []interface{}{
							"VPCID",
							"SubnetId",
							"SecurityGroupID",
						},
					},
				},
				"ParameterLabels": map[string]interface{}{
					"VPCID": map[string]interface{}{
						"default": "Which VPC should this be deployed to?",
					},
				},
			},
			"Foo": "bar",
		},
		"Transform": "AWS::Serverless",
	})

	src, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "ok to overwrite",
		"Description":              "Line 2",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterGroups": []interface{}{
					map[string]interface{}{
						"Label": map[string]interface{}{
							"default": "Amazon EC2 Configuration",
						},
						"Parameters": []interface{}{
							"InstanceType",
							"KeyName",
						},
					},
				},
				"ParameterLabels": map[string]interface{}{
					"KeyName": map[string]interface{}{
						"default": "EC2 Instance Ker Pair",
					},
				},
			},
			"Baz": "quux",
		},
		"Parameters": map[string]interface{}{
			"Name": map[string]interface{}{
				"Type": "String",
			},
		},
		"Transform": map[string]interface{}{
			"Name": "AWS::Include",
			"Parameters": map[string]interface{}{
				"Location": "Somewhere",
			},
		},
	})

	expected, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "ok to overwrite",
		"Description":              "Line 1\nLine 2",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterGroups": []interface{}{
					map[string]interface{}{
						"Label": map[string]interface{}{
							"default": "Network Configuration",
						},
						"Parameters": []interface{}{
							"VPCID",
							"SubnetId",
							"SecurityGroupID",
						},
					},
					map[string]interface{}{
						"Label": map[string]interface{}{
							"default": "Amazon EC2 Configuration",
						},
						"Parameters": []interface{}{
							"InstanceType",
							"KeyName",
						},
					},
				},
				"ParameterLabels": map[string]interface{}{
					"VPCID": map[string]interface{}{
						"default": "Which VPC should this be deployed to?",
					},
					"KeyName": map[string]interface{}{
						"default": "EC2 Instance Ker Pair",
					},
				},
			},
			"Foo": "bar",
			"Baz": "quux",
		},
		"Parameters": map[string]interface{}{
			"Name": map[string]interface{}{
				"Type": "String",
			},
		},
		"Transform": []interface{}{
			"AWS::Serverless",
			map[string]interface{}{
				"Name": "AWS::Include",
				"Parameters": map[string]interface{}{
					"Location": "Somewhere",
				},
			},
		},
	})

	forceMerge = false
	actual, err := mergeTemplates(dst, src)
	if err != nil {
		t.Fatal(err)
	}

	if d := cmp.Diff(actual.Map(), expected.Map()); d != "" {
		t.Errorf("%s", d)
	}
}

func TestForceMergeTemplatesSuccess(t *testing.T) {
	dst, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "overwritten",
		"Description":              "Line 1",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterLabels": map[string]interface{}{
					"VPCID": map[string]interface{}{
						"default": "Which VPC should this be deployed to?",
					},
				},
			},
			"Foo": "bar",
		},
		"Parameters": map[string]interface{}{
			"Name": map[string]interface{}{
				"Type": "String",
			},
		},
	})

	src, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "ok to overwrite",
		"Description":              "Line 2",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterLabels": map[string]interface{}{
					"VPCID": map[string]interface{}{
						"default": "Which VPC should this be deployed to?",
					},
				},
			},
			"Foo": "quux",
		},
		"Parameters": map[string]interface{}{
			"Name": map[string]interface{}{
				"Type": "String",
			},
		},
	})

	expected, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "ok to overwrite",
		"Description":              "Line 1\nLine 2",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterLabels": map[string]interface{}{
					"VPCID": map[string]interface{}{
						"default": "Which VPC should this be deployed to?",
					},
					"VPCID_2": map[string]interface{}{
						"default": "Which VPC should this be deployed to?",
					},
				},
			},
			"Foo":   "bar",
			"Foo_2": "quux",
		},
		"Parameters": map[string]interface{}{
			"Name": map[string]interface{}{
				"Type": "String",
			},
			"Name_2": map[string]interface{}{
				"Type": "String",
			},
		},
	})

	forceMerge = true
	actual, err := mergeTemplates(dst, src)
	if err != nil {
		t.Fatal(err)
	}

	if d := cmp.Diff(actual.Map(), expected.Map()); d != "" {
		t.Errorf("%s", d)
	}
}

func TestEmptyMergeTemplatesSuccess(t *testing.T) {
	src, _ := parse.Map(map[string]interface{}{
		"AWSTemplateFormatVersion": "ok to overwrite",
		"Description":              "Line 2",
		"Metadata": map[string]interface{}{
			"AWS::CloudFormation::Interface": map[string]interface{}{
				"ParameterGroups": []interface{}{
					map[string]interface{}{
						"Label": map[string]interface{}{
							"default": "Amazon EC2 Configuration",
						},
						"Parameters": []interface{}{
							"InstanceType",
							"KeyName",
						},
					},
				},
				"ParameterLabels": map[string]interface{}{
					"KeyName": map[string]interface{}{
						"default": "EC2 Instance Ker Pair",
					},
				},
			},

			"Baz": "quux",
		},
		"Parameters": map[string]interface{}{
			"Name": map[string]interface{}{
				"Type": "String",
			},
		},
		"Transform": map[string]interface{}{
			"Name": "AWS::Include",
			"Parameters": map[string]interface{}{
				"Location": "Somewhere",
			},
		},
		"Resources": map[string]interface{}{
			"Type": "AWS::SSM::Parameter",
			"Properties": map[string]interface{}{
				"Name":  "test",
				"Type":  "String",
				"Value": "Value",
			},
		},
	})

	empty, _ := parse.Map(map[string]interface{}{})

	forceMerge = false
	// rain merge src.yaml /dev/null
	{
		actual, err := mergeTemplates(src, empty)
		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual.Map(), src.Map()); d != "" {
			t.Errorf("%s", d)
		}
	}

	// rain merge /dev/null src.yaml
	{
		actual, err := mergeTemplates(empty, src)
		if err != nil {
			t.Fatal(err)
		}

		if d := cmp.Diff(actual.Map(), src.Map()); d != "" {
			t.Errorf("%s", d)
		}
	}

}

func TestMergeTemplatesClash(t *testing.T) {
	dst, _ := parse.Map(map[string]interface{}{
		"Description": "Line 1",
		"Metadata": map[string]interface{}{
			"Foo": "bar",
		},
	})

	src, _ := parse.Map(map[string]interface{}{
		"Description": "Line 2",
		"Metadata": map[string]interface{}{
			"Foo": "baz",
		},
	})

	forceMerge = false
	if _, err := mergeTemplates(dst, src); err == nil {
		t.Fail()
	}
}

// TestMergeOutputs tests merging templates where one has Fn::Import statements that reference Outputs in the other
func TestMergeOutputs(t *testing.T) {

	// Export a value
	t1 := `
Resources:
  Bucket:
    Type: AWS::S3::Bucket
Outputs:
  BucketName:
    Value: !Ref Bucket
    Export: 
      Name: BucketNameExport
`

	// Reference the value from a different template
	t2 := `
Resources:
  AccessLogsBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: 
        Fn::Sub:
          - ${ParentBucket}-access-logs
          - ParentBucket: 
              Fn::ImportValue: BucketNameExport
`

	// The merged template converts the Import to a Ref
	expected := `
Resources:
  Bucket:
    Type: AWS::S3::Bucket
  AccessLogsBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub
        - ${ParentBucket}-access-logs
        - ParentBucket: !Ref Bucket
Outputs:
  BucketName:
    Value: !Ref Bucket
    Export: 
      Name: BucketNameExport
`

	template1, err := parse.String(t1)
	if err != nil {
		t.Fatal(err)
	}

	template2, err := parse.String(t2)
	if err != nil {
		t.Fatal(err)
	}

	expectedTemplate, err := parse.String(expected)
	if err != nil {
		t.Fatal(err)
	}

	mergeImports = true
	merged, err := mergeTemplates(template1, template2)
	if err != nil {
		t.Fatal(err)
	}

	if d := diff.New(expectedTemplate, merged); d.Mode() != "=" {
		t.Errorf("%s", d.Format(true))
	}

	mergeImports = false
	merged, err = mergeTemplates(template1, template2)
	if err != nil {
		t.Fatal(err)
	}

	if d := diff.New(expectedTemplate, merged); d.Mode() == "=" {
		t.Errorf("%s", d.Format(true))
	}
}

// TestMergeParseError calls the handler directly so a panic cannot be hidden by
// the CLI's recovery wrapper.
func TestMergeParseError(t *testing.T) {
	oldUsage, oldErrors := Cmd.SilenceUsage, Cmd.SilenceErrors
	defer func() { Cmd.SilenceUsage, Cmd.SilenceErrors = oldUsage, oldErrors }()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.yaml")
	bad := filepath.Join(dir, "bad.yaml")
	for path, content := range map[string]string{good: "Resources: {}\n", bad: "Resources: [\n"} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{bad, good}, {good, bad}} {
		err := Cmd.RunE(Cmd, args)
		if err == nil {
			t.Fatal("expected parse error")
		}
		for _, want := range []string{bad, "invalid YAML:", "did not find expected node content"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q does not contain %q", err, want)
			}
		}
	}
}

func TestMergeCommandErrors(t *testing.T) {
	oldSilenceUsage, oldSilenceErrors := Cmd.SilenceUsage, Cmd.SilenceErrors
	oldDebug := config.Debug
	config.Debug = true
	defer func() {
		config.Debug = oldDebug
		Cmd.SilenceUsage, Cmd.SilenceErrors = oldSilenceUsage, oldSilenceErrors
		Cmd.SetArgs(nil)
		Cmd.SetOut(nil)
		Cmd.SetErr(nil)
	}()
	for _, tc := range []struct{ name, content, want string }{
		{"parse", "Resources: [\n", "invalid YAML:"},
		{"conflict", "Resources:\n  Bucket:\n    Type: AWS::S3::Bucket\n", "templates have clashing Resources: Bucket"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.yaml")
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			Cmd.SetOut(&stdout)
			Cmd.SetErr(&stderr)
			Cmd.SetArgs([]string{path, path})
			err := Cmd.Execute()
			if err == nil {
				t.Fatal("expected command error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.name == "parse" && !strings.Contains(err.Error(), path) {
				t.Fatalf("missing filename: %v", err)
			}
			if got := stderr.String(); got != err.Error()+"\n" {
				t.Errorf("stderr = %q; want one error line %q", got, err.Error()+"\n")
			}
			if strings.Contains(stderr.String(), "Usage:") || stdout.Len() != 0 {
				t.Errorf("unexpected command output: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestMergeArgumentErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"no arguments", []string{}, "requires at least 2 arg(s), only received 0"},
		{"one argument", []string{"a.yaml"}, "requires at least 2 arg(s), only received 1"},
		{"unknown flag", []string{"--bogus", "a", "b"}, "unknown flag: --bogus"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Use a fresh command, as each CLI invocation starts with fresh Cobra state.
			command := &cobra.Command{Use: Cmd.Use, Args: Cmd.Args, RunE: Cmd.RunE, SilenceUsage: Cmd.SilenceUsage, SilenceErrors: Cmd.SilenceErrors}
			var output bytes.Buffer
			command.SetOut(&output)
			command.SetErr(&output)
			command.SetArgs(tc.args)
			err := command.Execute()
			if err == nil || err.Error() != tc.want {
				t.Fatalf("error = %v; want %q", err, tc.want)
			}
			for _, want := range []string{"Error: " + tc.want + "\n", "Usage:"} {
				if !strings.Contains(output.String(), want) {
					t.Errorf("output %q does not contain %q", output.String(), want)
				}
			}
		})
	}
}
