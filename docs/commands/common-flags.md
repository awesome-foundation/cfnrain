# Common flags

These flags work the same way on every command that has them. The command pages list them, and link here for the details.

Every command prints its own flags with `--help`, for example `cfnrain deploy --help`.

## Every command

| Flag | What it does |
|------|--------------|
| `-h`, `--help` | Print the usage and the flags of the command |
| `--debug` | Print debug information |
| `--no-colour` | Do not use color in the output. Use it in CI logs. |

## AWS profile and region

Every command that calls AWS takes these flags. `diff`, `fmt`, `merge` and `tree` do not call AWS, so they do not have them.

| Flag | What it does |
|------|--------------|
| `-p`, `--profile` | The AWS profile to use, from the AWS CLI configuration file |
| `-r`, `--region` | The AWS region to use |

cfnrain uses the standard AWS SDK credential chain. See [Install](../install.md#aws-credentials).

## Artifact bucket

`deploy`, `pkg`, `bootstrap` and `stackset deploy` take these flags. The `cc` subcommands take `--s3-bucket` and `--s3-prefix`. See [The artifact bucket](../concepts/artifact-bucket.md).

| Flag | What it does |
|------|--------------|
| `--s3-bucket` | The bucket to upload to, instead of the default bucket |
| `--s3-prefix` | A prefix for the object keys |
| `--s3-owner` | The account that must own the bucket. Uploads to a bucket that another account owns fail. |

## Parameters and tags

`deploy`, `stackset deploy`, `cc deploy` and `forecast` take `--params`, `--tags` and `-c` (`--config`). See [Parameters and tags](../concepts/deploying.md#parameters-and-tags).

## Experimental

`-x` (`--experimental`) turns on an experimental feature. On `deploy`, `-x` means `--no-exec`. See [Experimental features](../concepts/experimental.md).
