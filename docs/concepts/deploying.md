# Deploying

This page explains what [`cfnrain deploy`](../commands/stack/deploy.md) does, where it gets parameter and tag values, and how it uses change sets. For a first walk-through, see [Your first deploy](../getting-started.md).

## The deploy flow

`cfnrain deploy <template> [stack]` does these steps:

1. It [packages](packaging.md) the template.
2. It finds the [artifact bucket](artifact-bucket.md). If the bucket does not exist, cfnrain asks to create it.
3. It picks the stack name. If you do not give one, the stack name is the template file name without its extension.
4. It asks for each parameter that does not have a value from another source. See [Parameters and tags](#parameters-and-tags). If the parameter has a default or the stack has an existing value, the prompt shows it, and an empty answer keeps it.
5. It creates a change set and shows it. `+` is an addition, `>` is a modification and `-` is a removal. Nested stacks show under their parent.

    ```text
    CloudFormation will make the following changes:
    Stack bucket:
      + AWS::S3::Bucket Bucket
    Do you wish to continue? (Y/n)
    ```

6. If you answer yes, it executes the change set, shows the progress, and prints a summary with the outputs and any failure messages.

If you answer no, cfnrain deletes the change set. If the stack is new, it also deletes the empty stack that CloudFormation created for the change set.

cfnrain always sends the `CAPABILITY_NAMED_IAM` and `CAPABILITY_AUTO_EXPAND` capabilities. Templates larger than 51,200 bytes go to the artifact bucket first. The limit for those is 1 MB.

## Updates

On an update, cfnrain asks only for parameters that have no value from the other sources. An empty answer keeps the existing value. cfnrain sends it to CloudFormation as `UsePreviousValue`.

If the template has no changes, cfnrain prints `Change set was created, but there is no change. Deploy was skipped.` and exits with status 0.

!!! warning "Failed stacks are deleted without a prompt"
    If the existing stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, cfnrain deletes it before it deploys. It does not ask first, even without `-y`. Stacks in these states have no working resources. If the stack is in any other state that does not end in `_COMPLETE`, cfnrain stops with an error.

## Without questions

`-y` (`--yes`) answers yes to every question:

- cfnrain does not ask for parameters. Every parameter needs a value from a flag, the config file, an environment variable, a default, or the existing stack.
- cfnrain does not show the change set before it executes it.
- cfnrain creates the artifact bucket if it does not exist.

If a parameter has no value, cfnrain stops with this error:

```text
no default or existing value for parameter 'Env'. Set a default, supply a --params flag, or deploy without the --yes flag
```

## Exit status

The exit status is 0 when the stack reaches `CREATE_COMPLETE`, `UPDATE_COMPLETE` or `IMPORT_COMPLETE`, and when there are no changes. Any other result exits with a status that is not 0.

## Parameters and tags

cfnrain gets parameter values and stack tags from four sources.

### Flags

`--params` sets parameter values and `--tags` sets stack tags. Both take `key=value` pairs separated by commas:

```sh
cfnrain deploy bucket.yaml --params Env=prod --tags Team=platform,CostCenter=1234
```

A part with no `=` belongs to the value before it. `--params Subnets=subnet-a,subnet-b` sets `Subnets` to `subnet-a,subnet-b`. This is how you pass a `CommaDelimitedList`.

### Config file

`-c` or `--config` reads parameters and tags from a YAML or JSON file:

```yaml
Parameters:
  Env: prod
Tags:
  Team: platform
  CostCenter: "1234"
```

```sh
cfnrain deploy bucket.yaml billing-prod -c prod.yaml
```

The file format is like the AWS CodePipeline template configuration file, without `StackPolicy`. cfnrain also reads the lowercase keys `parameters:` and `tags:`. To get a config file from a stack that is already deployed, run `cfnrain cat -c <stack>`.

### Environment variables

`RAIN_VAR_<Name>` sets the parameter `<Name>`. `RAIN_DEFAULT_TAG_<Key>` sets the tag `<Key>`:

```sh
export RAIN_VAR_Env=dev
export RAIN_DEFAULT_TAG_Team=platform
cfnrain deploy bucket.yaml -y
```

!!! warning "Unknown parameters stop the deploy"
    cfnrain treats every `RAIN_VAR_*` variable as a parameter. If the template has no parameter with that name, the deploy stops with `unknown parameter`. Add `--ignore-unknown-params` to print a warning and continue. The same rule applies to `--params` and the config file.

### Precedence

When a parameter or tag has values from more than one source, cfnrain uses the first one in this list:

1. `--params` and `--tags`
2. The `--config` file
3. `RAIN_VAR_*` and `RAIN_DEFAULT_TAG_*` environment variables
4. Your answer at the prompt. cfnrain asks only when no source above has a value, and only without `-y`.
5. The existing value on the stack, for updates
6. The parameter's `Default` in the template

When a flag overrides a value from the config file, cfnrain prints a yellow warning.

## Change sets

Every deploy uses a change set. `--no-exec` creates the change set, prints it, and stops:

```sh
cfnrain deploy bucket.yaml billing-prod --no-exec
```

```text
Changeset contains the following changes:
Stack billing-prod:
  > AWS::S3::Bucket Bucket
changeset created but not executed: billing-prod-1761480000
```

The default change set name is `<stack>-<unix time>`. Give a name as the third argument to choose it:

```sh
cfnrain deploy bucket.yaml billing-prod release-42 --no-exec
```

Execute the change set later:

```sh
cfnrain deploy --changeset billing-prod release-42
```

`cfnrain ls <stack>` lists the change sets of a stack. `ls -c` shows one change set and `rm -c` deletes one:

```sh
cfnrain ls billing-prod
cfnrain ls -c billing-prod release-42
cfnrain rm -c billing-prod release-42
```

On `deploy`, `-x` is short for `--no-exec`, not `--experimental`. See [Experimental features](experimental.md#-x-means-no-exec-on-deploy).
