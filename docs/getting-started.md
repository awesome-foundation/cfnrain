# Getting started

This page deploys a small stack, changes it, and removes it. It also explains how cfnrain gets parameter and tag values.

You need AWS credentials that can create CloudFormation stacks and S3 buckets. Run `cfnrain info` to see the account and region that cfnrain uses.

## Your first deploy

Save this template as `bucket.yaml`:

```yaml
Parameters:
  Env:
    Type: String
    Description: The environment name
    AllowedValues: [dev, prod]

Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags:
        - Key: Env
          Value: !Ref Env

Outputs:
  BucketName:
    Value: !Ref Bucket
```

Deploy it:

```sh
cfnrain deploy bucket.yaml
```

cfnrain does these steps:

1. It packages the template. Packaging processes [template directives](directives/index.md) and [modules](modules/index.md), and uploads local files to the artifact bucket.
2. It finds the artifact bucket. If the bucket does not exist, cfnrain asks to create it. See [Artifact bucket](#the-artifact-bucket).
3. It picks the stack name. If you do not give one, the stack name is the template file name without its extension: `bucket`.
4. It asks for each parameter that does not have a value from another source:

    ```text
    Enter a value for parameter 'Env' "The environment name":
    ```

    If the parameter has a default or the stack has an existing value, the prompt shows it, and an empty answer keeps it.

5. It creates a change set and shows it. `+` is an addition, `>` is a modification and `-` is a removal. Nested stacks show under their parent.

    ```text
    CloudFormation will make the following changes:
    Stack bucket:
      + AWS::S3::Bucket Bucket
    Do you wish to continue? (Y/n)
    ```

6. If you answer yes, it executes the change set, shows the progress, and prints a summary with the outputs and any failure messages.

If you answer no, cfnrain deletes the change set. If the stack is new, it also deletes the empty stack that CloudFormation created for the change set.

To use a different stack name, give it as the second argument:

```sh
cfnrain deploy bucket.yaml billing-dev
```

## Change the stack

Edit the template and run the same command again. cfnrain shows the new change set. If the template has no changes, cfnrain prints `Change set was created, but there is no change. Deploy was skipped.` and exits with status 0.

On an update, cfnrain asks only for parameters that have no value from the other sources. An empty answer keeps the existing value. cfnrain sends it to CloudFormation as `UsePreviousValue`.

!!! warning "Failed stacks are deleted without a prompt"
    If the existing stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, cfnrain deletes it before it deploys. It does not ask first, even without `-y`. Stacks in these states have no working resources. If the stack is in any other state that does not end in `_COMPLETE`, cfnrain stops with an error.

## Skip the questions with -y

`-y` (`--yes`) answers yes to every question. cfnrain does not ask for parameters and does not show the change set before it executes it:

```sh
cfnrain deploy bucket.yaml billing-dev -y --params Env=dev
```

With `-y`, every parameter needs a value from a flag, the config file, an environment variable, a default, or the existing stack. If one has no value, cfnrain stops with this error:

```text
no default or existing value for parameter 'Env'. Set a default, supply a --params flag, or deploy without the --yes flag
```

## Parameters and tags

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

## Change sets without executing

`--no-exec` creates the change set, prints it, and stops:

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

!!! note "On deploy, -x means --no-exec"
    On `pkg`, `forecast`, `cc` and `module`, `-x` is short for `--experimental`. On `deploy`, `-x` is short for `--no-exec`, and `--experimental` has no short form. `cfnrain deploy -x` creates a change set and does not execute it.

## Other deploy options

| Flag | What it does |
|------|--------------|
| `-d`, `--detach` | Start the deploy and exit. Watch it later with `cfnrain watch <stack>`. |
| `-k`, `--keep` | Keep the resources after a failure. This disables rollback. |
| `-t`, `--termination-protection` | Turn on termination protection after a successful deploy. There is no flag to turn it off. |
| `--role-arn` | The IAM role that CloudFormation assumes to deploy the stack. |
| `--nested-change-set=false` | Do not include nested stacks in the change set. |
| `--experimental` | Allow experimental features, such as [S3 bucket metadata](bucket-metadata.md) and `!Rain::Module`. |

cfnrain always sends the `CAPABILITY_NAMED_IAM` and `CAPABILITY_AUTO_EXPAND` capabilities. Templates larger than 51,200 bytes go to the artifact bucket first. The limit for those is 1 MB.

See the [deploy reference](cfnrain_deploy.md) for every flag.

## Remove the stack

```sh
cfnrain rm billing-dev
```

cfnrain shows the stack and asks you to confirm. Add `-y` to skip the question.

## The artifact bucket

cfnrain uploads packaged files and large templates to an S3 bucket. It finds the bucket in this order:

1. The `--s3-bucket` flag
2. The SSM parameter `rain-bucket`
3. `rain-artifacts-<account id>-<region>`

If the bucket does not exist, cfnrain asks to create it. With `-y` it creates it without a question. You can also create it with `cfnrain bootstrap`. cfnrain creates the bucket with AES256 encryption, blocks all public access, and adds a lifecycle rule that deletes objects after 7 days. The rule's ID says `delete after 14 days`, but the real value is 7 days. See [Known issues](known-issues.md).

`--s3-prefix` adds a prefix to the object keys. `--s3-owner` sets the account that must own the bucket. cfnrain sends the owner with every upload, so an upload to a bucket that another account owns fails.

cfnrain resolves the bucket on every deploy, also when the template has nothing to upload.
