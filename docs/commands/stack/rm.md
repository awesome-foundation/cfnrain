# rm

`cfnrain rm <stack>` deletes a stack and waits until the delete completes. It asks you to confirm first. `remove`, `del` and `delete` are aliases.

```sh
cfnrain rm app-dev
cfnrain rm app-dev -y
```

If the stack has termination protection, cfnrain asks whether to turn it off. With `-y` it turns it off without a question.

| Flag | What it does |
|------|--------------|
| `-y`, `--yes` | Do not ask questions |
| `-d`, `--detach` | Start the delete and exit |
| `--role-arn` | The IAM role that CloudFormation assumes to delete the stack |
| `-c`, `--changeset` | Delete a change set instead of a stack: `cfnrain rm -c <stack> <changeset>` |
| `--experimental` | Needed when a bucket in the stack has [`Metadata.Rain`](../../bucket-metadata.md) |
