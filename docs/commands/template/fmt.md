# fmt

`cfnrain fmt` formats templates. It sorts the sections and properties in a fixed order, uses the short form of most intrinsic functions (`!Sub`, `!Ref`), and keeps your YAML comments, including comments on function keys it shortens. `format` is an alias. For comments on functions inside lists, see [Known issues](../../known-issues.md#fmt-drops-a-line-comment-on-a-function-inside-a-list).

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

Ten functions keep the form that your template uses, short or long: `Fn::Contains`, `Fn::EachMemberEquals`, `Fn::EachMemberIn`, `Fn::RefAll`, `Fn::ValueOf`, `Fn::ValueOfAll`, `Fn::Length`, `Fn::ToJsonString`, `Fn::GetStackOutput` and `Fn::Transform`. The AWS template reference does not document a short form for most of them. Inside a short `!GetStackOutput`, every function keeps its form, because AWS does not allow a short `!Sub` or `!Ref` there. JSON output always uses the long form.

| Flag | What it does |
|------|--------------|
| `-w`, `--write` | Write the result back to the file |
| `-v`, `--verify` | Do not print. Exit with status 0 if the file is already formatted, and 1 if it is not. |
| `-j`, `--json` | Write JSON instead of YAML |
| `-u`, `--unsorted` | Keep the original order of sections and properties |
| `--uncdk` | Remove CDK metadata, simplify logical IDs, and turn `Fn::Join` into `Fn::Sub` |
| `--node-style` | Choose how to write YAML values: `tagged`, `doublequoted`, `singlequoted`, `literal`, `folded`, `strict-boolean`, `quotescalars`, `original` or `flow` |
| `-p`, `--pkl` | Write [Pkl](../../pkl.md) instead of YAML |
| `--pkl-basic` | With `--pkl`, do not use the Pkl modules of the `cfn` package |
| `--pkl-package` | With `--pkl`, the alias or full URI of the Pkl package. The default is `@cfn`. |
| `--datamodel` | Print the internal YAML data model. This is for debugging cfnrain. |

YAML output preserves merge keys (`<<: *anchor`). JSON output rejects merge keys with an error naming `<<`; expand the merged mapping before using `--json`. A quoted `"<<"` key is an ordinary key and can be written as JSON.

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
