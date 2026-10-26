# !Rain::Embed

`!Rain::Embed <path>` reads a file and inserts its contents into the template as a string. cfnrain removes the whitespace at the start and end of the file. The path must be a file, not a directory.

Use it for inline Lambda code, scripts, policies in a text format, and other text that is easier to edit in its own file.

## Example

The template:

```yaml
--8<-- "cft/pkg/tmpl/docs/embed-template.yaml"
```

`handler.py`, in the same directory:

```python
--8<-- "cft/pkg/tmpl/docs/handler.py"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/embed-expect.yaml"
```

## Related

- [`!Rain::Include`](include.md) parses the file as YAML or JSON and inserts an object, not a string.
- [`Fn::InsertFile`](../modules/intrinsics.md#fninsertfile) also inserts a file as a string. Unlike `!Rain::Embed`, it keeps the whitespace at the end.
