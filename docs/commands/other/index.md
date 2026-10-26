# Other commands

These commands do not work with templates or stacks directly. Both take `-p` (`--profile`) and `-r` (`--region`). Both also take `--debug` and `--no-colour`, and print their usage with `-h` (`--help`).

The output on these pages is illustrative. It needs a real AWS account.

| Command | What it does |
|---------|--------------|
| [`info`](info.md) | Shows the account, region, identity and profile that cfnrain uses. |
| [`console`](console.md) | Makes a sign-in URL for the AWS console from your current credentials and opens it in a browser. |
