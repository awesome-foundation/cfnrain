# deploy

`cfnrain deploy <template> [stack] [changeset]` creates or updates a stack through a change set. [Deploying](../../concepts/deploying.md) explains the flow, the parameter sources and change sets.

```sh
# Deploy, with a prompt for each parameter and a confirmation
cfnrain deploy app.yaml app-dev

# Deploy without questions, with values from a config file
cfnrain deploy app.yaml app-prod -y -c prod.yaml

# Create a change set and stop
cfnrain deploy app.yaml app-prod --no-exec

# Execute that change set later
cfnrain deploy --changeset app-prod app-prod-1761480000
```

If you do not give a stack name, the stack name is the template file name without its extension.

!!! warning "On deploy, -x means --no-exec"
    On other commands, `-x` means `--experimental`. See [Experimental features](../../concepts/experimental.md#-x-means-no-exec-on-deploy).

## Flags

| Flag | What it does |
|------|--------------|
| `-y`, `--yes` | Do not ask questions. See [Without questions](../../concepts/deploying.md#without-questions). |
| `--params` | Parameter values, in the format `key1=value1,key2=value2` |
| `--tags` | Stack tags, in the format `key1=value1,key2=value2` |
| `-c`, `--config` | A YAML or JSON file with parameters and tags. See [Config file](../../concepts/deploying.md#config-file). |
| `--ignore-unknown-params` | Warn about parameters that the template does not declare, and continue |
| `-x`, `--no-exec` | Create the change set and do not execute it. See [Change sets](../../concepts/deploying.md#change-sets). |
| `--changeset` | Execute an existing change set: `cfnrain deploy --changeset <stack> <changeset>` |
| `-d`, `--detach` | Start the deploy and exit. Watch it later with [`cfnrain watch`](watch.md). |
| `-k`, `--keep` | Keep the resources after a failure. This disables rollback. |
| `-t`, `--termination-protection` | Turn on termination protection after a successful deploy. There is no flag to turn it off. |
| `--role-arn` | The IAM role that CloudFormation assumes to deploy the stack |
| `--nested-change-set=false` | Do not include nested stacks in the change set |
| `--experimental` | Allow experimental features, such as [S3 bucket metadata](../../bucket-metadata.md) and `!Rain::Module` |
| `--node-style` | How to write YAML values in the template that cfnrain sends. The values are the same as for [`fmt --node-style`](../template/fmt.md). The default is `original`. |
| `--no-analytics` | Has no effect. cfnrain keeps it so that old scripts still work. |
| `--s3-bucket`, `--s3-prefix`, `--s3-owner` | See [Artifact bucket](../common-flags.md#artifact-bucket) |
| `-p`, `--profile`, `-r`, `--region` | See [AWS profile and region](../common-flags.md#aws-profile-and-region) |
