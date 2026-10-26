# watch

`cfnrain watch <stack>` shows the status of a stack while it changes. Use it for a deploy that another tool started, or after `deploy --detach`.

```sh
cfnrain deploy app.yaml app-prod -y --detach
cfnrain watch app-prod
```

If the stack is not changing, `watch` stops with an error. Add `-w` (`--wait`) to wait until a change starts.
