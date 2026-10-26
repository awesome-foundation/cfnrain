# Packaging

cfnrain adds features to the CloudFormation template language. CloudFormation does not know these features, so cfnrain turns them into a plain template before it sends the template to AWS. This step is packaging.

Packaging processes:

- [Template directives](../directives/index.md), such as `!Rain::Embed` and `!Rain::S3`
- [Constants](../constants.md)
- [Modules](../modules/index.md), including `ForEach`, conditions and the extra intrinsic functions
- [Implicit uploads](../directives/implicit-uploads.md): local paths in resource properties that take an S3 location

Files that packaging uploads go to the [artifact bucket](artifact-bucket.md).

## Which commands package

These commands package the template before they use it: [`pkg`](../commands/template/pkg.md), [`deploy`](../commands/stack/deploy.md), [`forecast`](../commands/template/forecast.md), [`stackset deploy`](../commands/stack/stackset.md#stackset-deploy) and [`cc deploy`](../commands/stack/cc.md).

`cfnrain pkg` packages a template and prints the result. Use it to see what CloudFormation gets:

```sh
cfnrain pkg app.yaml -o app.packaged.yaml
```

Run linters such as `cfn-lint` on the packaged template, because they do not know the cfnrain features.

## Paths

A path in a template is relative to the directory of the template file that contains it. It is not relative to the directory where you run cfnrain. In an included file or a module, paths are relative to that file.

## Templates can run code

Two features run programs from your computer:

- `Run` in [`!Rain::S3`](../directives/s3.md#run) runs a script when cfnrain packages the template.
- `RunBefore` and `RunAfter` in [S3 bucket metadata](../bucket-metadata.md) run commands when cfnrain deploys the template.

`Run` works without `--experimental`. Only package and deploy templates that you trust.
