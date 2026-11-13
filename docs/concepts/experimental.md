# Experimental features

Some features are experimental. They work, but their format or behavior can change in a minor release. You must turn each one on with `--experimental`, so you do not use one by accident.

| Feature | How to turn it on |
|---------|-------------------|
| **Deprecated.** [`cc`](../commands/stack/cc.md): deploy with the Cloud Control API. This command will be removed in cfnrain 2.0. See [Direction](https://github.com/awesome-foundation/cfnrain/discussions/20). | `-x` or `--experimental` on every `cc` command |
| **Deprecated.** [`forecast`](../commands/template/forecast.md): predict deploy failures. This command will be removed in cfnrain 2.0. See [Direction](https://github.com/awesome-foundation/cfnrain/discussions/20). | `-x` or `--experimental` |
| [`module`](../commands/template/module.md): modules in CodeArtifact | `-x` or `--experimental` on every `module` command |
| [`!Rain::Module`](../modules/packages.md#the-older-rainmodule-directive), and packages in the `Rain:` section | `-x` on `pkg`, `--experimental` on `deploy` |
| [S3 bucket metadata](../bucket-metadata.md) | `--experimental` on `deploy` and `rm` |

The [`Modules` section](../modules/index.md) is not experimental and needs no flag.

## -x means --no-exec on deploy

On `pkg`, `forecast`, `cc` and `module`, `-x` is short for `--experimental`. On `deploy`, `-x` is short for `--no-exec`, and `--experimental` has no short form. `cfnrain deploy -x` creates a change set and does not execute it.

In scripts, write the long flag names, so the intent is clear.
