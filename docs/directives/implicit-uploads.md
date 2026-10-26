# Implicit uploads

Some resource properties take an S3 location. If the template has a local path in one of these properties, cfnrain uploads the file and replaces the path with the S3 location. You do not need a directive. This is what `aws cloudformation package` does too.

!!! note "Illustrative output"
    Uploads need AWS, so the output on this page is illustrative.

## Example

```yaml
Resources:
  Function:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: python3.13
      Handler: index.main
      Role: !GetAtt FunctionRole.Arn
      Code: lambda-src

  Network:
    Type: AWS::CloudFormation::Stack
    Properties:
      TemplateURL: network.yaml
```

The packaged template:

```yaml
Resources:
  Function:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: python3.13
      Handler: index.main
      Role: !GetAtt FunctionRole.Arn
      Code:
        S3Bucket: rain-artifacts-123456789012-eu-west-1
        S3Key: 1b4844dacc843f09941c11c94f80981d3be8ae7578952c71e875ef7add37b1a7

  Network:
    Type: AWS::CloudFormation::Stack
    Properties:
      TemplateURL: https://rain-artifacts-123456789012-eu-west-1.s3.eu-west-1.amazonaws.com/d2a84f4b8b650937ec8f73cd8be2c74add5a911ba64df27458ed8229da804a26
```

## Properties

| Resource type | Property | Output |
|---------------|----------|--------|
| `AWS::ApiGateway::RestApi` | `BodyS3Location` | Object with `Bucket` and `Key` |
| `AWS::AppSync::FunctionConfiguration` | `RequestMappingTemplateS3Location`, `ResponseMappingTemplateS3Location` | S3 URI |
| `AWS::AppSync::GraphQLSchema` | `DefinitionS3Location` | S3 URI |
| `AWS::AppSync::Resolver` | `RequestMappingTemplateS3Location`, `ResponseMappingTemplateS3Location` | S3 URI |
| `AWS::CloudFormation::Stack` | `TemplateURL` | HTTPS URL. cfnrain packages the nested template first. |
| `AWS::ElasticBeanstalk::ApplicationVersion` | `SourceBundle` | Object with `S3Bucket` and `S3Key` |
| `AWS::Glue::Job` | `Command.ScriptLocation` | S3 URI |
| `AWS::Lambda::Function` | `Code` | Object with `S3Bucket` and `S3Key`, always zipped |
| `AWS::Lambda::LayerVersion` | `Content` | Object with `S3Bucket` and `S3Key`, always zipped |
| `AWS::Serverless::Api` | `DefinitionUri` | S3 URI |
| `AWS::Serverless::Application` | `Location` | HTTPS URL. cfnrain packages the nested template first. |
| `AWS::Serverless::Function` | `CodeUri` | S3 URI, always zipped |
| `AWS::Serverless::LayerVersion` | `ContentUri` | S3 URI, always zipped |
| `AWS::ServerlessRepo::Application` | `LicenseUrl`, `ReadmeUrl` | S3 URI |
| `AWS::StepFunctions::StateMachine` | `DefinitionS3Location` | Object with `Bucket` and `Key` |

## Rules

- cfnrain only changes a value that is a plain string. A map, such as an existing `S3Bucket` and `S3Key`, stays as it is.
- An S3 URI (`s3://...`) stays as it is. For the nested stack properties, an `http://` or `https://` URL stays as it is.
- If the upload fails, for example because the path does not exist, cfnrain leaves the value as it is and shows no error. Run with `--debug` to see why. The nested stack properties are different: a missing nested template stops packaging.
- Nested templates are packaged with all their directives and modules, then uploaded.

The AWS CLI also packages `AWS::Include` `Location`, `AWS::Serverless::StateMachine` `DefinitionUri` and `AWS::CodeCommit::Repository` `Code.S3`. cfnrain does not. Use [`!Rain::S3`](s3.md) for those.
