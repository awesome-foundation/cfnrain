package cft

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

// ParseOnlyTags are short tags that the parser expands but fmt never writes.
// The AWS template reference does not document most of them as short forms.
// !GetStackOutput forbids the short form around !Sub and !Ref, which fmt writes,
// and writing !Transform would drop a YAML anchor on its mapping.
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
