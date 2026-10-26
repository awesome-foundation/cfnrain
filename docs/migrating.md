# Migrating from rain

cfnrain starts from Rain v1.24.4 plus 22 later commits. Templates, config files, environment variables and the artifact bucket work as before. Most migrations only change the name of the binary.

## Steps

1. [Install cfnrain](install.md).
2. Replace `rain` with `cfnrain` in your scripts and CI workflows. Or link the old name, so the scripts do not change:

    ```sh
    ln -s "$(command -v cfnrain)" /usr/local/bin/rain
    ```

3. If an IAM role trust policy limits the session name, allow `cfnrain-*`. See [Role session name](#role-session-name).
4. Remove the Rain binary or the Homebrew formula when you no longer need it.

## What stays the same

| Item | Status |
|------|--------|
| `!Rain::*` directives | Same |
| The `Rain:` template section | Same |
| The `Modules`, `Packages` and `Constants` sections | Same |
| `Metadata.Rain` on S3 buckets | Same |
| Config files for `--config` | Same |
| `RAIN_VAR_*` and `RAIN_DEFAULT_TAG_*` | Same |
| Artifact bucket `rain-artifacts-<account>-<region>` | Same bucket, cfnrain uses it as it is |
| SSM parameter `rain-bucket` | Same |
| `cc` state files in the artifact bucket | Same location |
| CodeArtifact defaults for `module` (`--domain cloudformation`, `--repo rain`) | Same |
| Flags and commands | Same |

## What changed

### Binary and module name

The command is `cfnrain`. The Go module is `github.com/awesome-foundation/cfnrain`. `go install` uses the new path:

```sh
go install github.com/awesome-foundation/cfnrain/cmd/cfnrain@latest
```

`cfnrain --version` prints `cfnrain v<version> <os>/<arch>`. Scripts that parse the version output need an update.

### User agent

AWS sees the user agent `cfnrain/<version>`, not `Rain/<version>`. If you filter CloudTrail events or write IAM conditions on the user agent, update them.

### Role session name

When cfnrain assumes a role from your AWS profile, the role session name is `cfnrain-<version>`, not `Rain-<version>`. If a role trust policy only allows session names that match `Rain-*`, add `cfnrain-*`:

```json
"Condition": {
  "StringLike": {
    "sts:RoleSessionName": ["Rain-*", "cfnrain-*"]
  }
}
```

### No analytics metadata

Rain added a `Metadata.AWSToolsMetrics.Rain` block to every packaged template. cfnrain does not. The `--no-analytics` flag is still accepted, and has no effect. If you compare packaged templates from Rain and cfnrain, expect this block to be the difference.

### Documentation

The Rain documentation at `aws-cloudformation.github.io/rain` is frozen. This site replaces it. Some things that the old README said were wrong:

- On `deploy`, `-x` is `--no-exec`, not `--experimental`. See [Getting started](getting-started.md#change-sets-without-executing).
- The `Modules` section does not need `--experimental`. Only `!Rain::Module` and bucket metadata do.
- `!Rain::S3` takes `Format: URI`, not `Uri`. See [S3](directives/s3.md#format).
- The `ForEach` placeholders are `$Identifier` and `$Index`. There is no `$MapValue`. See [ForEach](modules/foreach.md).
- Packaged resources get no `SourceMap` metadata.
- `!Rain::Module` does not need a `ModuleExtension` resource or `Metadata.Extends`.

See [Known issues](known-issues.md) for behavior that is still wrong in the code.
