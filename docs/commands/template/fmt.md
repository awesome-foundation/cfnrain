# fmt

`cfnrain fmt` formats templates. It sorts the sections and properties in a fixed order, uses the short form of most intrinsic functions (`!Sub`, `!Ref`), and keeps your YAML comments, including comments on function keys it shortens. `format` is an alias.

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

A function whose argument has a YAML anchor keeps its long form, so aliases still refer to the argument rather than the whole function.

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

The `singlequoted`, `doublequoted`, and `quotescalars` node styles quote strings. They also quote string keys such as `Integer`. Numbers, booleans, and nulls stay plain and keep their types. Strings such as `'0'`, `'1.0'`, and `'true'` stay quoted strings. The `singlequoted` style uses double quotes for strings with line breaks to preserve their exact text. The `quotescalars` style changes only scalars. Lists and mappings keep the flow or block style that the other formatting rules choose.

The `literal` and `folded` node styles apply block formatting only to strings. Numbers, booleans, and nulls keep their types.

With no file names, `fmt` reads standard input.

JSON output preserves empty function arguments as `""`. A bare short-form function tag with no value also becomes `""`, including tags from `pkg` output. This does not change `~` or `null` arguments.

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

The `flow` node style uses flow mappings even when `fmt` sorts their keys. The `quotescalars` style preserves flow mappings during sorting.
