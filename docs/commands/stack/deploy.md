# deploy

`cfnrain deploy <template> [stack] [changeset]` creates or updates a stack through a change set. [Getting started](../../getting-started.md) explains the full flow, parameters, config files and change sets.

```sh
# Deploy, with a prompt for each parameter and a confirmation
cfnrain deploy app.yaml app-dev

# Deploy without questions
cfnrain deploy app.yaml app-prod -y -c prod.yaml

# Create a change set and stop
cfnrain deploy app.yaml app-prod --no-exec
```

[Getting started](../../getting-started.md#other-deploy-options) lists every `deploy` flag.
