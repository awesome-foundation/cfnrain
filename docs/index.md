# cfnrain

cfnrain is a command line tool for AWS CloudFormation templates and stacks. It deploys stacks through change sets, packages local files into templates, formats templates, and shows stack logs.

cfnrain is a community fork of [AWS CloudFormation Rain](https://github.com/aws-cloudformation/rain), which AWS archived on 2026-07-29. It works as a replacement for Rain v1.24.4. See [Migrating from rain](migrating.md).

!!! note "Not an AWS project"
    cfnrain is not affiliated with or endorsed by Amazon Web Services.

## What cfnrain does

- **Deploys stacks with change sets.** [`cfnrain deploy`](commands/stack/deploy.md) asks for parameter values, shows the change set, and waits for your confirmation. Then it shows the progress and prints a summary with any failure messages.
- **Adds features to templates.** [Template directives](directives/index.md) such as `!Rain::Embed` and `!Rain::S3` put file contents or S3 locations into a template. [Modules](modules/index.md) split one template into many files.
- **Formats templates.** [`cfnrain fmt`](commands/template/fmt.md) writes templates in one consistent style, keeps your comments, and converts between JSON and YAML.
- **Shows stack logs.** [`cfnrain logs`](commands/stack/logs.md) shows the events of a stack and its nested stacks, and hides the events that are not interesting.

## Where to start

1. [Install cfnrain](install.md).
2. Do [your first deploy](getting-started.md).
3. Read [how deploys work](concepts/deploying.md).
4. Use cfnrain [in CI](ci.md).

## Project

- Source and releases: [github.com/awesome-foundation/cfnrain](https://github.com/awesome-foundation/cfnrain)
- Questions and ideas: [GitHub Discussions](https://github.com/awesome-foundation/cfnrain/discussions)
- License: Apache 2.0. The `NOTICE` file keeps the original Rain copyright.
