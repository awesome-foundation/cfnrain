# !Rain::Env

`!Rain::Env <name>` inserts the value of an environment variable into the template as a string. If the variable is not set, packaging stops with the error `missing environmental variable`.

Use it for values that come from the build system, such as a commit ID or the name of the pipeline that deploys the stack.

## Example

The template:

```yaml
--8<-- "cft/pkg/tmpl/docs/env-template.yaml"
```

The packaged template, with `DEPLOYED_BY=ci-pipeline` in the environment:

```yaml
--8<-- "cft/pkg/tmpl/docs/env-expect.yaml"
```

!!! note "Parameters are often a better fit"
    The value is fixed in the packaged template. CloudFormation does not see where it came from. For values that change between environments, a template parameter with [`RAIN_VAR_<Name>`](../getting-started.md#environment-variables) shows the value in the stack parameters.
