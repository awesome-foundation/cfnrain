# cfnrain

cfnrain is a command line tool for AWS CloudFormation templates and stacks. It deploys stacks through change sets, packages local files into templates, formats templates, and shows stack logs.

cfnrain is a community fork of [AWS CloudFormation Rain](https://github.com/aws-cloudformation/rain). AWS archived Rain on 2026-07-29 and plans no more releases. cfnrain continues from Rain's last commit, which is Rain v1.24.4 plus 22 commits.

!!! note "Not an AWS project"
    cfnrain is not affiliated with or endorsed by Amazon Web Services.

## What cfnrain does

- **Deploys stacks with change sets.** [`cfnrain deploy`](getting-started.md) packages the template, asks for parameter values, shows the change set, and waits for your confirmation. Then it streams the progress and prints a summary with any failure messages.
- **Packages local files into templates.** [Template directives](directives/index.md) such as `!Rain::Embed` and `!Rain::S3` put file contents or S3 locations into a template. [Modules](modules/index.md) split one template into many files.
- **Formats templates.** [`cfnrain fmt`](commands/template.md#fmt) writes templates in one consistent style, keeps your comments, and converts between JSON and YAML.
- **Shows stack logs.** [`cfnrain logs`](commands/stack.md#logs) shows the events of a stack and its nested stacks, and hides the events that are not interesting.

## Compatibility with Rain

cfnrain works as a replacement for Rain v1.24.4. Your templates, config files and AWS accounts work with no changes:

- The `!Rain::*` directives, the `Rain:` template section, the `Modules`, `Packages` and `Constants` sections, and the bucket `Metadata.Rain` commands work as before.
- The `RAIN_VAR_*` and `RAIN_DEFAULT_TAG_*` environment variables work as before.
- cfnrain uses the same artifact bucket (`rain-artifacts-<account>-<region>`) and the same `rain-bucket` SSM parameter.
- Scripts that call `rain` keep working when you link the name. See [Install](install.md#keep-the-rain-name).

## What changed

| Item | Rain | cfnrain |
|------|------|---------|
| Command name | `rain` | `cfnrain` |
| Go module | `github.com/aws-cloudformation/rain` | `github.com/awesome-foundation/cfnrain` |
| User agent that AWS sees | `Rain/<version>` | `cfnrain/<version>` |
| Role session name when cfnrain assumes a role from your profile | `Rain-<version>` | `cfnrain-<version>` |
| `Metadata.AWSToolsMetrics` block in packaged templates | Added | Not added. `--no-analytics` is accepted and has no effect. |

If a role trust policy allows only session names that match `Rain-*`, add `cfnrain-*`. See [Migrating from rain](migrating.md) for the full list.

## Where to start

1. [Install cfnrain](install.md).
2. [Deploy your first stack](getting-started.md).
3. Read about the [commands](commands/stack.md) and the [template directives](directives/index.md).
4. Use cfnrain [in CI](ci.md).

The [command reference](cfnrain.md) lists every command and flag. It is generated from the code, so it always matches the binary.

## Project

- Source and releases: [github.com/awesome-foundation/cfnrain](https://github.com/awesome-foundation/cfnrain)
- Questions and ideas: [GitHub Discussions](https://github.com/awesome-foundation/cfnrain/discussions)
- License: Apache 2.0. The `NOTICE` file keeps the original Rain copyright.
