package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-foundation/cfnrain/cft"
	"github.com/awesome-foundation/cfnrain/cft/parse"
	"github.com/awesome-foundation/cfnrain/cft/visitor"
	"github.com/awesome-foundation/cfnrain/internal/config"
	"github.com/awesome-foundation/cfnrain/internal/node"
	"github.com/awesome-foundation/cfnrain/internal/s11n"
	"gopkg.in/yaml.v3"
)

// plainSubConstantError is deferred until the containing constant is used.
type plainSubConstantError struct {
	error
}

// rainConstant parses a !Rain::Constant node
// Constants can be any type of YAML node, but if the constant is a string, it
// can be used in a Sub with ${Rain::ConstantName}. Otherwise use a directive.
// !Rain::Constant ConstantName. Constants are evaluated in order, so they can
// refer to other constants declared previously.
func rainConstant(ctx *directiveContext) (bool, error) {

	name := ctx.n.Content[1].Value
	if err := ctx.t.ConstantErrors[name]; err != nil {
		return false, err
	}
	val, ok := ctx.t.Constants[name]
	if !ok {
		return false, fmt.Errorf("rain constant %s not found", name)
	}

	ctx.t.Origins.Copy(ctx.n, val)
	*ctx.n = *val

	return true, nil
}

// replaceConstants replaces ${Rain::Name} and ${Const::Name} in a single
// scalar node. If the constant name is not found in the map created from the
// Rain section In the template, an error is returned
func replaceConstants(n *yaml.Node, constants map[string]*yaml.Node, constantErrors map[string]error, inSub bool) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected n to be a ScalarNode")
	}

	// Parse every scalar as if it was a Sub. Look for ${Rain::X}

	retval := ""
	words, err := parse.ParseSub(n.Value, true)
	if err != nil {
		// Preserve legacy behavior for literal strings and substitutions that
		// ParseSub cannot parse. Only constant resolution errors are fatal.
		config.Debugf("replaceConstants skipped: %v", err)
		return nil
	}
	var deferredErr error
	for _, w := range words {
		switch w.T {
		case parse.STR:
			retval += w.W
		case parse.REF:
			retval += fmt.Sprintf("${%s}", w.W)
		case parse.AWS:
			retval += fmt.Sprintf("${AWS::%s}", w.W)
		case parse.GETATT:
			retval += fmt.Sprintf("${%s}", w.W)
		case parse.RAIN:
			if err := constantErrors[w.W]; err != nil {
				var plainSubErr *plainSubConstantError
				if !errors.As(err, &plainSubErr) {
					return err
				}
				if deferredErr == nil {
					deferredErr = err
				}
				continue
			}
			val, ok := constants[w.W]
			if !ok {
				return fmt.Errorf("did not find Constant %s", w.W)
			}
			// A string-form Sub is stored as a mapping, so insert its string
			// argument and leave any CloudFormation variables for the outer Sub.
			if val.Kind == yaml.MappingNode && len(val.Content) == 2 &&
				val.Content[0].Value == "Fn::Sub" && val.Content[1].Kind == yaml.ScalarNode {
				if !inSub && isSubNeeded(val.Content[1].Value) {
					if deferredErr == nil {
						deferredErr = &plainSubConstantError{fmt.Errorf("constant %s uses Fn::Sub and cannot be interpolated into a plain string. Use !Sub for the outer value or !Rain::Constant %s for the whole value", w.W, w.W)}
					}
					continue
				}
				val = val.Content[1]
			}
			if val.Kind != yaml.ScalarNode {
				return fmt.Errorf("constant %s cannot be interpolated as a string; use !Rain::Constant or !Ref Const::%s to insert the whole value", w.W, w.W)
			}
			config.Debugf("Found a Rain constant: %s = %s", w.W, val.Value)
			retval += val.Value
		}
	}

	// Keep the original value when any substitution is invalid.
	if deferredErr != nil {
		return deferredErr
	}
	if n.Value != retval {
		n.Value = retval
	}

	return nil
}

// replaceTemplateConstants scans the entire template looking for Sub strings
// and replaces all instances of ${Rain::Name} and ${Const::Name} if that name
// exists in the Rain/Constants section of the template
func replaceTemplateConstants(t *cft.Template, definitions *yaml.Node) error {

	constants := t.Constants

	for k, v := range constants {
		config.Debugf("replaceTemplateConstants %s: %v", k, node.YamlStr(v))
	}

	invalidDefinitions := make(map[*yaml.Node]bool)
	for name := range t.ConstantErrors {
		invalidDefinitions[constants[name]] = true
	}
	var err error

	vf := func(v *visitor.Visitor) {
		if err != nil {
			return
		}
		yamlNode := v.GetYamlNode()
		if v.GetParentNode() == definitions && invalidDefinitions[yamlNode] {
			v.SkipChildren()
			return
		}
		if yamlNode.Kind == yaml.MappingNode && len(yamlNode.Content) == 2 {

			key := yamlNode.Content[0].Value
			switch key {
			case "Rain::Constant":
				name := yamlNode.Content[1].Value
				if val, ok := constants[name]; ok {
					if err = t.ConstantErrors[name]; err != nil {
						return
					}
					t.Origins.Copy(yamlNode, val)
					*yamlNode = *node.Clone(val, t.Origins)
				}
			case "Fn::Sub":
				arg := yamlNode.Content[1]
				if arg.Kind == yaml.SequenceNode && len(arg.Content) == 2 {
					arg = arg.Content[0]
				}
				if arg.Kind != yaml.ScalarNode {
					return
				}
				err = replaceConstants(arg, constants, t.ConstantErrors, true)
				if err != nil {
					config.Debugf("%v", err)
					return
				}

				// Remove unnecessary Subs
				// Keep escaped variables for CloudFormation to resolve as literals.
				if yamlNode.Content[1].Kind == yaml.ScalarNode && !isSubNeeded(arg.Value) {
					*yamlNode = *node.MakeScalar(arg.Value)
				}
			case "Ref":
				// Ref is used with object constants
				val := yamlNode.Content[1]
				if strings.HasPrefix(val.Value, parse.CONSTcc) {
					constName := strings.Replace(val.Value, parse.CONSTcc, "", 1)
					if err = t.ConstantErrors[constName]; err != nil {
						return
					}
					constVal, ok := constants[constName]
					if !ok {
						err = fmt.Errorf("constant %s not found", constName)
						return
					}
					t.Origins.Copy(yamlNode, constVal)
					*yamlNode = *node.Clone(constVal, t.Origins)
				}
			}
		}
	}

	visitor := visitor.NewVisitor(t.Node)
	visitor.Visit(vf)

	return err
}

func processConstants(t *cft.Template, n *yaml.Node) error {
	if t.ConstantErrors == nil {
		t.ConstantErrors = make(map[string]error)
	}
	// Process constants in order, since they can refer to previous ones
	_, c, _ := s11n.GetMapValue(n, string(cft.Constants))
	if c != nil {
		for i := 0; i < len(c.Content); i += 2 {
			name := c.Content[i].Value
			val := c.Content[i+1]
			t.Constants[name] = val
			delete(t.ConstantErrors, name)
			// Visit each node in val looking for prior constant entries
			var err, deferredErr error
			subStrings := make(map[*yaml.Node]bool)
			vf := func(v *visitor.Visitor) {
				if err != nil {
					return
				}
				// Keep the first deferred error, but check later nodes for hard errors.
				defer func() {
					var plainSubErr *plainSubConstantError
					if errors.As(err, &plainSubErr) {
						if deferredErr == nil {
							deferredErr = err
						}
						err = nil
					}
				}()
				vnode := v.GetYamlNode()
				if vnode.Kind == yaml.MappingNode && len(vnode.Content) == 2 {
					key, value := vnode.Content[0].Value, vnode.Content[1].Value
					if key == "Rain::Constant" {
						err = t.ConstantErrors[value]
					} else if key == "Ref" && strings.HasPrefix(value, parse.CONSTcc) {
						err = t.ConstantErrors[strings.TrimPrefix(value, parse.CONSTcc)]
					}
					if err != nil {
						return
					}
				}
				if vnode.Kind == yaml.MappingNode && len(vnode.Content) == 2 && vnode.Content[0].Value == "Fn::Sub" {
					arg := vnode.Content[1]
					if arg.Kind == yaml.SequenceNode && len(arg.Content) == 2 {
						arg = arg.Content[0]
					}
					subStrings[arg] = true
				}
				if vnode.Kind == yaml.ScalarNode {
					err = replaceConstants(vnode, t.Constants, t.ConstantErrors, subStrings[vnode])
				}
			}
			v := visitor.NewVisitor(val)
			v.Visit(vf)
			if err != nil {
				return fmt.Errorf("constant %s: %w", name, err)
			}
			if deferredErr != nil {
				t.ConstantErrors[name] = fmt.Errorf("constant %s: %w", name, deferredErr)
			}
		}
	}

	// Now scan the template to replace constants
	err := replaceTemplateConstants(t, c)
	if err != nil {
		return err
	}
	node.RemoveFromMap(n, string(cft.Constants))
	return nil
}
