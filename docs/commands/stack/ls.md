# ls

`cfnrain ls` lists the stacks in the current region. Add `-a` (`--all`) to list the stacks in all regions.

```sh
cfnrain ls
cfnrain ls -a
```

Give a stack name to see a summary of that stack and its change sets. With `-a`, the summary has more detail.

```sh
cfnrain ls app-prod
```

`-c` (`--changeset`) works with change sets. With no arguments it lists the stacks that have change sets. With a stack and a change set name it shows that change set.

```sh
cfnrain ls -c
cfnrain ls -c app-prod release-42
```

`list` is an alias for `ls`.
