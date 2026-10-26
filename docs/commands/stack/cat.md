# cat

`cfnrain cat <stack>` prints the template of a deployed stack. cfnrain formats it like [`fmt`](../template/fmt.md) does.

```sh
cfnrain cat app-prod
```

| Flag | What it prints |
|------|----------------|
| `-t`, `--transformed` | The template after CloudFormation processed transforms such as `AWS::Serverless-2016-10-31` |
| `-u`, `--unformatted` | The template as CloudFormation stores it, with no formatting |
| `-c`, `--config` | A config file with the stack's parameters and tags, in the format that `deploy --config` reads |

Use `-c` to copy the settings of one stack to a new deploy:

```sh
cfnrain cat -c app-prod > prod.yaml
cfnrain deploy app.yaml app-prod -c prod.yaml
```
