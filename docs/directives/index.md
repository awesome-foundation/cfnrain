# Template directives

Directives are YAML tags that start with `!Rain::`. cfnrain processes them when it packages a template, and replaces each one with a plain value. CloudFormation never sees them.

See [Packaging](../concepts/packaging.md) for which commands package a template. Run `cfnrain pkg` to see the packaged template before you deploy it.

| Directive | What it inserts | Needs AWS |
|-----------|-----------------|-----------|
| [`!Rain::Embed <path>`](embed.md) | The contents of a file, as a string | No |
| [`!Rain::Include <path>`](include.md) | A YAML or JSON file, as an object | No |
| [`!Rain::Env <name>`](env.md) | The value of an environment variable, as a string | No |
| [`!Rain::S3 <path>`](s3.md) | The S3 URI of an uploaded file, or an object with the bucket and key | Yes |
| [`!Rain::S3Http <path>`](s3.md#rains3http) | The HTTPS URL of an uploaded file | Yes |
| [`!Rain::Constant <name>`](../constants.md#the-rain-section) | The value of a constant from the `Rain:` section | No |
| [`!Rain::Module <uri>`](../modules/packages.md#the-older-rainmodule-directive) | The resources of a module. This is the older module system and needs `--experimental`. | No |

cfnrain also uploads local files for some resource properties without a directive, like `aws cloudformation package` does. See [Implicit uploads](implicit-uploads.md).

## Paths

A path in a directive is relative to the template file that contains it. See [Paths](../concepts/packaging.md#paths).

## Long form

The YAML tag is a short form. The long form is a map with one key, and it also works in JSON templates:

```json
{
  "Code": {
    "ZipFile": { "Rain::Embed": "handler.py" }
  }
}
```

## Examples come from tests

The examples on these pages are test files in [`cft/pkg/tmpl/docs`](https://github.com/awesome-foundation/cfnrain/tree/main/cft/pkg/tmpl/docs). `go test ./cft/pkg/` packages each `-template.yaml` file and compares the result with its `-expect.yaml` file. If cfnrain changes how a directive works, the test fails. The S3 directives need AWS, so their examples are illustrative and not tested.
