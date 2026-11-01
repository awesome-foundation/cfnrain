package cft

import "gopkg.in/yaml.v3"

// ShortForm is a yaml.Node Style bit that the parser sets on the mapping it
// makes from a short tag, so fmt can write the form the source used.
// yaml.v3 tests only its own Style bits, so it never writes this one.
const ShortForm yaml.Style = 1 << 31

// Tags is a mapping from YAML short tags to full instrincic function names
var Tags = map[string]string{
	"!And":            "Fn::And",
	"!Base64":         "Fn::Base64",
	"!Cidr":           "Fn::Cidr",
	"!Equals":         "Fn::Equals",
	"!FindInMap":      "Fn::FindInMap",
	"!GetAZs":         "Fn::GetAZs",
	"!GetAtt":         "Fn::GetAtt",
	"!If":             "Fn::If",
	"!ImportValue":    "Fn::ImportValue",
	"!Join":           "Fn::Join",
	"!Not":            "Fn::Not",
	"!Or":             "Fn::Or",
	"!Select":         "Fn::Select",
	"!Split":          "Fn::Split",
	"!Sub":            "Fn::Sub",
	"!Ref":            "Ref",
	"!Condition":      "Condition",
	"!Rain::Embed":    "Rain::Embed",
	"!Rain::Include":  "Rain::Include",
	"!Rain::Env":      "Rain::Env",
	"!Rain::S3Http":   "Rain::S3Http",
	"!Rain::S3":       "Rain::S3",
	"!Rain::Module":   "Rain::Module",
	"!Rain::Constant": "Rain::Constant",
	"!Invoke":         "Fn::Invoke",
	"!InsertFile":     "Fn::InsertFile",
	"!Merge":          "Fn::Merge",
}

// ParseOnlyTags are short tags that fmt writes only where the source used them.
// The AWS template reference does not document most of them as short forms,
// and !GetStackOutput forbids the short form around !Sub and !Ref.
var ParseOnlyTags = map[string]string{
	"!Contains":         "Fn::Contains",
	"!EachMemberEquals": "Fn::EachMemberEquals",
	"!EachMemberIn":     "Fn::EachMemberIn",
	"!GetStackOutput":   "Fn::GetStackOutput",
	"!Length":           "Fn::Length",
	"!RefAll":           "Fn::RefAll",
	"!ToJsonString":     "Fn::ToJsonString",
	"!Transform":        "Fn::Transform",
	"!ValueOf":          "Fn::ValueOf",
	"!ValueOfAll":       "Fn::ValueOfAll",
}
