# Stack commands

These commands work with deployed CloudFormation stacks. All of them take `-p` (`--profile`) and `-r` (`--region`). Every command also takes `--debug` and `--no-colour`, and prints its usage with `-h` (`--help`).

The command output on these pages is illustrative. It needs a real AWS account, so the names, IDs and times in your output will be different.

| Command | What it does |
|---------|--------------|
| [`deploy`](deploy.md) | Creates or updates a stack through a change set. |
| [`ls`](ls.md) | Lists the stacks in the current region. |
| [`cat`](cat.md) | Prints the template of a deployed stack. |
| [`logs`](logs.md) | Shows the events of a stack and its nested stacks. |
| [`watch`](watch.md) | Shows the status of a stack while it changes. |
| [`rm`](rm.md) | Deletes a stack and waits until the delete completes. |
| [`stackset`](stackset.md) | Works with CloudFormation stack sets. |
| [`cc`](cc.md) | Deploys templates with the Cloud Control API instead of CloudFormation. Experimental. |
