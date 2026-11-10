package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awesome-foundation/cfnrain/cft/format"
	"gopkg.in/yaml.v3"
)

func init() {
	registry["Resources/*|Type==AWS::ApiGateway::RestApi/Properties/BodyS3Location"] = wrapObject("Bucket", "Key", false)
	registry["Resources/*|Type==AWS::AppSync::FunctionConfiguration/Properties/RequestMappingTemplateS3Location"] = wrapS3URI
	registry["Resources/*|Type==AWS::AppSync::FunctionConfiguration/Properties/ResponseMappingTemplateS3Location"] = wrapS3URI
	registry["Resources/*|Type==AWS::AppSync::GraphQLSchema/Properties/DefinitionS3Location"] = wrapS3URI
	registry["Resources/*|Type==AWS::AppSync::Resolver/Properties/RequestMappingTemplateS3Location"] = wrapS3URI
	registry["Resources/*|Type==AWS::AppSync::Resolver/Properties/ResponseMappingTemplateS3Location"] = wrapS3URI
	registry["Resources/*|Type==AWS::CloudFormation::Stack/Properties/TemplateURL"] = wrapTemplate
	registry["Resources/*|Type==AWS::ElasticBeanstalk::ApplicationVersion/Properties/SourceBundle"] = wrapObject("S3Bucket", "S3Key", false)
	registry["Resources/*|Type==AWS::Glue::Job/Properties/Command/ScriptLocation"] = wrapS3URI
	registry["Resources/*|Type==AWS::Lambda::Function/Properties/Code"] = wrapObject("S3Bucket", "S3Key", true)
	registry["Resources/*|Type==AWS::Lambda::LayerVersion/Properties/Content"] = wrapObject("S3Bucket", "S3Key", true)
	registry["Resources/*|Type==AWS::Serverless::Api/Properties/DefinitionUri"] = wrapS3URI
	registry["Resources/*|Type==AWS::Serverless::Application/Properties/Location"] = wrapTemplate
	registry["Resources/*|Type==AWS::Serverless::Function/Properties/CodeUri"] = wrapS3ZipURI
	registry["Resources/*|Type==AWS::Serverless::LayerVersion/Properties/ContentUri"] = wrapS3ZipURI
	registry["Resources/*|Type==AWS::ServerlessRepo::Application/Properties/LicenseUrl"] = wrapS3URI
	registry["Resources/*|Type==AWS::ServerlessRepo::Application/Properties/ReadmeUrl"] = wrapS3URI
	registry["Resources/*|Type==AWS::StepFunctions::StateMachine/Properties/DefinitionS3Location"] = wrapObject("Bucket", "Key", false)
}

func wrapS3(n *yaml.Node, root string, options s3Options) (bool, error) {
	if strings.HasPrefix(options.Path, "s3://") || strings.HasPrefix(options.Path, "http://") || strings.HasPrefix(options.Path, "https://") {
		return false, nil
	}
	newNode, err := handleS3(root, options)
	if err != nil {
		return false, fmt.Errorf("implicit upload %q at line %d, column %d: %w", options.Path, n.Line, n.Column, err)
	}

	replaceNode(n, newNode)

	return true, nil
}

func wrapS3ZipURI(ctx *directiveContext) (bool, error) {
	n := ctx.n
	root := ctx.rootDir
	if n.Kind != yaml.ScalarNode {
		// No need to error - this could be valid
		return false, nil
	}

	return wrapS3(n, root, s3Options{
		Path:   n.Value,
		Format: s3URI,
		Zip:    true,
	})
}

func wrapS3URI(ctx *directiveContext) (bool, error) {
	n := ctx.n
	if n.Kind != yaml.ScalarNode {
		// No need to error - this could be valid
		return false, nil
	}

	if strings.HasPrefix(n.Value, "s3://") {
		return false, nil // Already an s3 uri
	}

	return wrapS3(n, ctx.rootDir, s3Options{
		Path:   n.Value,
		Format: s3URI,
	})
}

func wrapObject(bucket, key string, forceZip bool) directiveFunc {
	return func(ctx *directiveContext) (bool, error) {
		n := ctx.n
		root := ctx.rootDir
		if n.Kind != yaml.ScalarNode {
			// No need to error - this could be valid
			return false, nil
		}

		return wrapS3(n, root, s3Options{
			Path:           n.Value,
			Format:         s3Object,
			BucketProperty: bucket,
			KeyProperty:    key,
			Zip:            forceZip,
		})
	}
}

func wrapTemplate(ctx *directiveContext) (bool, error) {
	n := ctx.n
	root := ctx.rootDir
	if n.Kind != yaml.ScalarNode {
		// No need to error - this could be valid
		return false, nil
	}

	if strings.HasPrefix(n.Value, "s3://") || strings.HasPrefix(n.Value, "http://") || strings.HasPrefix(n.Value, "https://") {
		return false, nil // Already http
	}

	path := n.Value
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}

	tmpl, err := File(path)
	if err != nil {
		return false, fmt.Errorf("nested template %q: %w", path, err)
	}

	f, err := os.CreateTemp(os.TempDir(), "*.template")
	if err != nil {
		return false, err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write([]byte(format.String(tmpl, format.Options{}))); err != nil {
		return false, err
	}

	if err := f.Close(); err != nil {
		return false, err
	}

	return wrapS3(n, root, s3Options{
		Path:   f.Name(),
		Format: s3Http,
	})
}
