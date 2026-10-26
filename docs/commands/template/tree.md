# tree

`cfnrain tree <template>` shows which parameters, resources and outputs depend on each other. `graph` is an alias.

```sh
cfnrain tree bucket.yaml
```

```yaml
Resources:
  Bucket:
    DependsOn:
      Parameters:
        - AWS::StackName
Outputs:
  Name:
    DependsOn:
      Resources:
        - Bucket
```

`-b` (`--both`) also shows what depends on each item. `-a` (`--all`) shows items with no dependencies. `-d` (`--dot`) writes the graph in GraphViz DOT format:

```sh
cfnrain tree -d bucket.yaml | dot -Tsvg > bucket.svg
```
