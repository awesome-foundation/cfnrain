# Template commands

These commands work with template files on your computer. `diff`, `fmt`, `merge` and `tree` need no AWS credentials. `pkg`, `bootstrap`, `build`, `forecast` and `module` can call AWS, so they take `-p` (`--profile`) and `-r` (`--region`). Every command also takes `--debug` and `--no-colour`, and prints its usage with `-h` (`--help`).

The output on these pages comes from real runs of cfnrain, unless a page says otherwise.

| Command | What it does |
|---------|--------------|
| [`fmt`](fmt.md) | Formats templates. |
| [`pkg`](pkg.md) | Packages a template. |
| [`diff`](diff.md) | Compares two template files. |
| [`merge`](merge.md) | Merges two or more templates into one. |
| [`tree`](tree.md) | Shows which parameters, resources and outputs depend on each other. |
| [`build`](build.md) | Makes starter templates from the CloudFormation resource schemas. |
| [`bootstrap`](bootstrap.md) | Creates the artifact bucket. |
| [`forecast`](forecast.md) | Looks for things that can make a deploy fail, such as a bucket name that already exists or an EC2 key pair that does not exist. Experimental. |
| [`module`](module.md) | Publishes and installs packages of modules with AWS CodeArtifact. Experimental. |
