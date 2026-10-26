# !Rain::Include

`!Rain::Include <path>` reads a YAML or JSON file and inserts it into the template as an object. cfnrain also processes the directives in the included file. Paths in the included file are relative to that file.

Use it to share tags, policies, or whole resource definitions between templates.

## Example

The template:

```yaml
--8<-- "cft/pkg/tmpl/docs/include-template.yaml"
```

`tags.yaml`, in the same directory:

```yaml
--8<-- "cft/pkg/tmpl/docs/tags.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/include-expect.yaml"
```

## Related

- [`!Rain::Embed`](embed.md) inserts a file as a string.
- [Modules](../modules/index.md) insert a group of resources, with parameters and outputs.
