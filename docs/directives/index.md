# Template directives

Directives are YAML tags that start with `!Rain::`. cfnrain processes them when it packages a template, and replaces each one with a plain value. CloudFormation never sees them.

These commands package templates: [`pkg`](../commands/template/pkg.md), [`deploy`](../getting-started.md), [`forecast`](../commands/template/forecast.md), [`stackset deploy`](../commands/stack/stackset.md#stackset-deploy) and [`cc deploy`](../commands/stack/cc.md). Run `cfnrain pkg` to see the packaged template before you deploy it.

| Directive | What it inserts | Needs AWS |
|-----------|-----------------|-----------|
| [`!Rain::Embed <path>`](embed.md) | The contents of a file, as a string | No |
| [`!Rain::Include <path>`](include.md) | A YAML or JSON file, as an object | No |
| [`!Rain::Env <name>`](env.md) | The value of an environment variable, as a string | No |
| [`!Rain::S3 <path>`](s3.md) | The S3 URI of an uploaded file, or an object with the bucket and key | Yes |
| [`!Rain::S3Http <path>`](s3.md#rains3http) | The HTTPS URL of an uploaded file | Yes |
| [`!Rain::Constant <name>`](constants.md) | The value of a constant from the `Rain:` section | No |
| [`!Rain::Module <uri>`](../modules/packages.md#the-older-rainmodule-directive) | The resources of a module. This is the older module system and needs `--experimental`. | No |

cfnrain also uploads local files for some resource properties without a directive, like `aws cloudformation package` does. See [Implicit uploads](implicit-uploads.md).

## Paths

A path in a directive is relative to the directory of the template file that contains it. It is not relative to the directory where you run cfnrain. In an included file or a module, paths are relative to that file.

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
