# Template commands

These commands work with template files on your computer. `diff`, `fmt`, `merge` and `tree` need no AWS credentials. `pkg`, `bootstrap`, `build`, `forecast` and `module` can call AWS, so they take `-p` (`--profile`) and `-r` (`--region`).

The output on this page comes from real runs of cfnrain, unless a section says otherwise.

## fmt

`cfnrain fmt` formats templates. It sorts the sections and properties in a fixed order, uses the short form of intrinsic functions (`!Sub`, `!Ref`), and keeps your YAML comments. `format` is an alias.

Input, `bucket.json`:

```json
{"Resources":{"Bucket":{"Properties":{"BucketName":{"Fn::Sub":"${AWS::StackName}-data"}},"Type":"AWS::S3::Bucket"}},"Outputs":{"Name":{"Value":{"Ref":"Bucket"}}}}
```

```sh
cfnrain fmt bucket.json
```

Output:

```yaml
Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub ${AWS::StackName}-data

Outputs:
  Name:
    Value: !Ref Bucket
```

| Flag | What it does |
|------|--------------|
| `-w`, `--write` | Write the result back to the file |
| `-v`, `--verify` | Do not print. Exit with status 0 if the file is already formatted, and 1 if it is not. |
| `-j`, `--json` | Write JSON instead of YAML |
| `-u`, `--unsorted` | Keep the original order of sections and properties |
| `--uncdk` | Remove CDK metadata, simplify logical IDs, and turn `Fn::Join` into `Fn::Sub` |
| `--node-style` | Choose how to write YAML values, for example `quotescalars` or `strict-boolean` |
| `-p`, `--pkl` | Write [Pkl](../pkl.md) instead of YAML |

With no file names, `fmt` reads standard input.

Format every template in place:

```sh
cfnrain fmt -w templates/*.yaml
```

Fail a CI job when a template is not formatted:

```sh
cfnrain fmt -v templates/app.yaml
```

```text
templates/app.yaml: would reformat
```

`--verify` compares the input text with the formatted YAML. A JSON file never passes unless you also give `-j`.

Reference: [cfnrain fmt](../cfnrain_fmt.md)

## pkg

`cfnrain pkg <template>` packages a template. It processes the [template directives](../directives/index.md), [modules](../modules/index.md) and [implicit uploads](../directives/implicit-uploads.md), and prints the result. It is like `aws cloudformation package`, with more features. `package` is an alias.

```sh
cfnrain pkg app.yaml -o app.packaged.yaml
```

`deploy` and `forecast` package the template too, so you only need `pkg` to see or keep the packaged template. Run linters such as `cfn-lint` on the packaged template, because they do not know the cfnrain directives.

| Flag | What it does |
|------|--------------|
| `-o`, `--output` | Write the result to a file |
| `-x`, `--experimental` | Allow `!Rain::Module` |
| `--s3-bucket`, `--s3-prefix`, `--s3-owner` | Control where uploads go. See [Artifact bucket](../getting-started.md#the-artifact-bucket). |

Reference: [cfnrain pkg](../cfnrain_pkg.md)

## diff

`cfnrain diff <from> <to>` compares two template files.

```sh
cfnrain diff old.yaml new.yaml
```

```text
(|) Resources:
(|)   Bucket:
(|)     Properties:
(|)       BucketName:
(>)         Fn::Sub: ${AWS::StackName}-files
(+)   Queue:
(+)     Type: AWS::SQS::Queue
```

`(+)` is an addition, `(-)` is a removal, `(>)` is a change and `(|)` is a parent of a change. `-l` (`--long`) also shows the parts that did not change. `diff` compares two local files. It does not compare a file with a deployed stack.

Reference: [cfnrain diff](../cfnrain_diff.md)

## merge

`cfnrain merge` merges two or more templates into one.

```sh
cfnrain merge bucket.yaml queue.yaml
```

```yaml
Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub ${AWS::StackName}-data

  Queue:
    Type: AWS::SQS::Queue
```

If two templates use the same logical ID, `merge` stops with an error. `-f` (`--force`) renames the second one instead. It does not update the `Ref`, `GetAtt` and `Sub` references to the renamed item. `--merge-imports` replaces an `Fn::ImportValue` with a direct reference when the export comes from another template in the merge. `-o` writes the result to a file.

Reference: [cfnrain merge](../cfnrain_merge.md)

## tree

`cfnrain tree <template>` shows which parameters, resources and outputs depend on each other. `graph` is an alias.

```sh
cfnrain tree bucket.yaml
```

```yaml
Resources:
  Bucket:
    DependsOn:
      Parameters:
        - AWS::StackName
Outputs:
  Name:
    DependsOn:
      Resources:
        - Bucket
```

`-b` (`--both`) also shows what depends on each item. `-a` (`--all`) shows items with no dependencies. `-d` (`--dot`) writes the graph in GraphViz DOT format:

```sh
cfnrain tree -d bucket.yaml | dot -Tsvg > bucket.svg
```

Reference: [cfnrain tree](../cfnrain_tree.md)

## build

`cfnrain build` makes starter templates from the CloudFormation resource schemas. The schemas are in the binary, so most uses need no AWS call.

Make a template with only the required properties of a type:

```sh
cfnrain build -b AWS::SQS::Queue
```

```yaml
AWSTemplateFormatVersion: 2010-09-09

Description: Generated by rain

Resources:
  Queue:
    Type: AWS::SQS::Queue
    Properties: {}
```

Without `-b`, the template has every property that you can set. You can give more than one type.

List the resource types that start with a prefix:

```sh
cfnrain build -l AWS::SQS
```

```text
AWS::SQS::Queue
AWS::SQS::QueueInlinePolicy
AWS::SQS::QueuePolicy
```

| Flag | What it does |
|------|--------------|
| `-b`, `--bare` | Only the required properties |
| `-j`, `--json` | JSON instead of YAML |
| `-s`, `--schema` | Print the registry schema of a type |
| `-l`, `--list` | List resource types, with an optional prefix |
| `-o`, `--output` | Write to a file |
| `--recommend` | Print a recommended template for a use case. With no more arguments, cfnrain shows a menu. `cfnrain build --recommend bucket bucket` prints the recommended bucket template. |
| `--pkl-class` | Print a [Pkl](../pkl.md) class for a type |
| `--prompt` | Write a template with Amazon Bedrock from a text prompt. The default model is `claude2`, which is old. Bedrock usage costs money. |

With no arguments, `build` shows an interactive menu.

!!! warning "--only-cache does the same as --no-cache"
    In the code, `--only-cache` sets the same switch as `--no-cache`. Both make cfnrain skip the cached schemas. See [Known issues](../known-issues.md).

Reference: [cfnrain build](../cfnrain_build.md)

## bootstrap

`cfnrain bootstrap` creates the artifact bucket. `deploy` and `pkg` create it too when they need it, so you only need `bootstrap` to create it ahead of time.

```sh
cfnrain bootstrap
cfnrain bootstrap -y --s3-bucket my-artifacts
```

See [Artifact bucket](../getting-started.md#the-artifact-bucket) for the bucket name and settings.

Reference: [cfnrain bootstrap](../cfnrain_bootstrap.md)

## forecast (experimental)

`cfnrain forecast` looks for things that can make a deploy fail, such as a bucket name that already exists or an EC2 key pair that does not exist. It is not a linter. Use `cfn-lint` to check the syntax. It needs `-x` (`--experimental`).

```sh
cfnrain forecast -x app.yaml app-prod
```

The output in this example is illustrative. When a check fails, it looks like this:

```text
Stormy weather ahead! 🌪

1 checks failed out of 9 total checks
FG001 FAIL on line 6: AWS::S3::Bucket Bucket - Resource with this name already exists
```

When all checks pass, the output starts with `Clear skies!` and the number of checks. The exit status is 1 when a check fails, so you can use `forecast` in CI.

| Flag | What it does |
|------|--------------|
| `--action` | `create`, `update`, `delete` or `all` (default) |
| `-a`, `--all` | Show all checks, not only the failed ones |
| `--include-iam` | Also check IAM permissions. This is slow. |
| `--role-arn` | The role to use for the IAM checks |
| `--type` | Check only one resource type |
| `--ignore` | Skip resource types or check codes, for example `AWS::S3::Bucket,F0002` |
| `--params`, `--tags`, `-c` | Parameter and tag values, as for `deploy` |

The checks cover S3 buckets and bucket policies, EC2 instances, launch templates and security groups, RDS clusters, load balancer listeners and target groups, SNS topics, Lambda functions and SageMaker notebook instances.

Reference: [cfnrain forecast](../cfnrain_forecast.md)

## module (experimental)

`cfnrain module` publishes and installs packages of modules with AWS CodeArtifact. See [Packages](../modules/packages.md#codeartifact-modules-experimental).

Reference: [cfnrain module](../cfnrain_module.md)
