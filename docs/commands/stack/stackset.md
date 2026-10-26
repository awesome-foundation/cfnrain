# stackset

`cfnrain stackset` works with CloudFormation stack sets. It has three subcommands. All of them take `--admin` to use delegated administrator permissions.

## stackset deploy

`cfnrain stackset deploy <template> [stackset]` creates or updates a stack set. If you do not give a name, the stack set name is the template file name without its extension.

```sh
cfnrain stackset deploy guardrails.yaml org-guardrails \
  --accounts 111111111111,222222222222 \
  --regions us-east-1,eu-west-1 \
  --params Env=prod
```

The config file can set the stack set and the instances, as well as parameters and tags:

```yaml
Parameters:
  Env: prod
Tags:
  Team: platform
StackSet:
  description: Organization guardrails
StackSetInstances:
  accounts:
    - "111111111111"
  regions:
    - us-east-1
```

Accounts and regions from the flags replace the values from the file. Tags and parameters from the flags are merged with the values from the file. `-i` (`--ignore-stack-instances`) updates the stack set and does not add or remove instances. See [test-config.yaml](https://github.com/awesome-foundation/cfnrain/blob/main/test/samples/test-config.yaml) for a full config file.

Like `deploy`, `stackset deploy` packages the template first, and creates the artifact bucket when it needs one.

| Flag | What it does |
|------|--------------|
| `--accounts` | The accounts to create stack set instances in, separated by commas |
| `--regions` | The regions to create stack set instances in, separated by commas |
| `--params` | Parameter values, in the format `key1=value1,key2=value2` |
| `--tags` | Tags, in the format `key1=value1,key2=value2` |
| `-c`, `--config` | A YAML or JSON config file, as shown above |
| `-i`, `--ignore-stack-instances` | Update the stack set and do not add or remove instances |
| `-y`, `--yes` | Do not ask for confirmation |
| `-d`, `--detach` | Start the operation and exit |
| `--admin` | Use delegated administrator permissions |
| `--s3-bucket`, `--s3-prefix`, `--s3-owner` | Control where uploads go. See [Artifact bucket](../common-flags.md#artifact-bucket). |

## stackset ls

`cfnrain stackset ls` lists the stack sets in the region. Give a name to see its instances and its last 10 operations.

```sh
cfnrain stackset ls
cfnrain stackset ls org-guardrails
```

`-a` (`--all`) lists the stack sets in all regions. With a name, it shows more detail.

## stackset rm

`cfnrain stackset rm <stackset>` deletes a stack set. If the stack set still has instances, cfnrain lists them and asks which ones to delete. Enter `0` to delete all of them.

```sh
cfnrain stackset rm org-guardrails
```

`-d` (`--detach`) starts the delete and exits.
