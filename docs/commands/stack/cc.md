# cc (experimental)

`cfnrain cc` deploys templates with the Cloud Control API instead of CloudFormation. Every `cc` command needs `-x` (`--experimental`).

```sh
cfnrain cc deploy -x app.yaml app-dev
cfnrain cc state -x app-dev
cfnrain cc drift -x app-dev
cfnrain cc rm -x app-dev
```

| Subcommand | What it does |
|------------|--------------|
| `cc deploy <template> <name>` | Creates or updates the resources. It refuses resource types that the Cloud Control API does not fully support. |
| `cc state <name>` | Prints the state file |
| `cc drift <name>` | Compares the state file with the live resources, and offers to fix the difference |
| `cc rm <name>` | Deletes the resources and waits until all the Cloud Control API calls complete |

`cc` keeps a state file in the artifact bucket at `deployments/<name>.yaml`. If a deploy stops and leaves a lock, `cc deploy -u <lockid>` removes the lock.

| Flag | Subcommands | What it does |
|------|-------------|--------------|
| `-x`, `--experimental` | all | Confirm that you know `cc` is experimental. Required. |
| `--s3-bucket`, `--s3-prefix` | all | The bucket and prefix of the state file and the uploads |
| `--s3-owner` | `cc` | The account that owns the bucket |
| `--params` | `deploy` | Parameter values, in the format `key1=value1,key2=value2` |
| `--tags` | `deploy` | Tags, in the format `key1=value1,key2=value2` |
| `-c`, `--config` | `deploy` | A YAML or JSON file with parameters and tags, as for [`deploy`](../../concepts/deploying.md#config-file) |
| `--ignore-unknown-params` | `deploy` | Ignore parameters that the template does not declare |
| `-u`, `--unlock` | `deploy` | Remove the lock with this ID and continue |
| `-y`, `--yes` | `deploy`, `rm` | Do not ask questions |

!!! warning "State files expire in buckets that cfnrain creates"
    The `cc` state files expire with the other objects in the artifact bucket. See [Objects expire after 7 days](../../concepts/artifact-bucket.md#objects-expire-after-7-days).
