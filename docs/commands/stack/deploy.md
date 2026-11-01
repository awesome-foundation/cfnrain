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

## Deleting an existing stack

If an existing stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, deploy asks before deleting it to make way for the new deployment. The prompt names the stack and its status and defaults to **No**. Declining stops deployment without deleting the stack. Deleting a stack also deletes its change sets, including ones created by someone else with `--no-exec` or the console.

With `-y` (`--yes`), deploy deletes the stack without asking, as before. This also applies when creating a change set with `--no-exec`.

The prompt uses the same confirmation helper as other commands: it requires stdout to be a terminal, and input errors stop the command before deletion. It does not automatically consent when stdin is redirected. Use `-y` for unattended deployment.

After deletion, deploy waits using the original stack ID, so it can track completion even when the stack name no longer resolves.

## Deployment result

After a change set executes, deploy waits for completion unless `--detach` is set. It then fetches the final stack summary. Deploy does not wait with `--no-exec` or when the change set has no changes.

If the final lookup fails after a successful wait, deploy still enables requested termination protection and runs post-deploy metadata commands. It then reports the lookup error with the stack name and exits unsuccessfully. It prints no stack summary or deployment success message.

## Flags

| Flag | What it does |
|------|--------------|
| `-y`, `--yes` | Do not ask questions. See [Without questions](../../concepts/deploying.md#without-questions). |
| `--params` | Parameter values, in the format `key1=value1,key2=value2` |
| `--tags` | Stack tags, in the format `key1=value1,key2=value2` |
| `-c`, `--config` | A YAML or JSON file with parameters and tags. See [Config file](../../concepts/deploying.md#config-file). |
| `--ignore-unknown-params` | Ignore parameters that the template does not declare, and continue without a warning |
| `-x`, `--no-exec` | Create the change set and do not execute it. See [Change sets](../../concepts/deploying.md#change-sets). |
| `--changeset` | Execute an existing change set: `cfnrain deploy --changeset <stack> <changeset>` |
| `-d`, `--detach` | Start the deploy and exit. Watch it later with [`cfnrain watch`](watch.md). |
| `-k`, `--keep` | Keep the resources after a failure. This disables rollback. |
| `-t`, `--termination-protection` | Turn on termination protection after a successful deploy. There is no flag to turn it off. |
| `--role-arn` | The IAM role that CloudFormation assumes to deploy the stack |
| `--nested-change-set=false` | Do not include nested stacks in the change set |
| `--express` | Use CloudFormation Express mode. Off by default. See [Express mode](#express-mode). |
| `--experimental` | Allow experimental features, such as [S3 bucket metadata](../../bucket-metadata.md) and `!Rain::Module` |
| `--node-style` | How to write YAML values in the template that cfnrain sends. The values are the same as for [`fmt --node-style`](../template/fmt.md). The default is `original`. |
| `--no-analytics` | Has no effect. cfnrain keeps it so that old scripts still work. |
| `--s3-bucket`, `--s3-prefix`, `--s3-owner` | See [Artifact bucket](../common-flags.md#artifact-bucket) |
| `-p`, `--profile`, `-r`, `--region` | See [AWS profile and region](../common-flags.md#aws-profile-and-region) |

## Express mode

With `--express`, the deploy finishes when CloudFormation applies the resource configuration. It does not wait for the resources to be ready to serve traffic. Use it for development, where speed is more important than a fully stable stack. AWS describes it in [Deploy stacks faster with express mode](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-express-mode.html).

```sh
cfnrain deploy app.yaml app-dev --express
```

- cfnrain sets Express mode on the change set. It applies to nested stacks too.
- Express mode in CloudFormation disables rollback by default. cfnrain keeps its own rule: it rolls back after a failure unless you also give `--keep`.
- With `--no-exec`, the change set keeps the Express mode setting. A later `--changeset` run executes it in Express mode. The change set also records the rollback setting, so give `--keep` on both runs or on neither.
- You cannot use `--express` with `--changeset`, because the mode is set when CloudFormation creates the change set.
- Custom resources still wait for their response, as in a standard deploy.

## When the change set fails

If CloudFormation [Early Validation](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/validate-stack-deployments.html) stops the change set, cfnrain gets the validation events with the `DescribeEvents` API. The error then shows each failed resource, the validation name and the reason.

The identity that runs cfnrain needs the `cloudformation:DescribeEvents` permission for this. Without it, cfnrain shows the original CloudFormation message.
