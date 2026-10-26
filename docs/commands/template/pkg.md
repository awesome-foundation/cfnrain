# pkg

`cfnrain pkg <template>` packages a template. It processes the [template directives](../../directives/index.md), [modules](../../modules/index.md) and [implicit uploads](../../directives/implicit-uploads.md), and prints the result. It is like `aws cloudformation package`, with more features. `package` is an alias.

```sh
cfnrain pkg app.yaml -o app.packaged.yaml
```

`deploy` and `forecast` package the template too, so you only need `pkg` to see or keep the packaged template. Run linters such as `cfn-lint` on the packaged template, because they do not know the cfnrain directives.

| Flag | What it does |
|------|--------------|
| `-o`, `--output` | Write the result to a file |
| `-x`, `--experimental` | Allow `!Rain::Module` |
| `--s3-bucket`, `--s3-prefix`, `--s3-owner` | Control where uploads go. See [Artifact bucket](../../getting-started.md#the-artifact-bucket). |
| `--node-style` | How to write YAML values, as for [`fmt`](fmt.md) |
| `--datamodel` | Print the internal YAML data model. This is for debugging cfnrain. |
| `--no-analytics` | Has no effect. cfnrain keeps it so that old scripts still work. |
