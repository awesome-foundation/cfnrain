# Known issues

These are places where cfnrain does something unexpected. They come from Rain, and cfnrain has not fixed them yet. Each one says what to do instead, and links to its issue. When a fix ships, its entry goes away.

## fmt drops a line comment on a function inside a list

When a list item is a function with sequence arguments, `fmt` drops a line comment on the function key. Examples include `Fn::Join` in a list and `Fn::Split` inside `Fn::Select`. Put the comment above the property that contains the list instead. See [#85](https://github.com/awesome-foundation/cfnrain/issues/85).

## build --only-cache does the same as --no-cache

`cfnrain build --only-cache` should use only the cached schemas. In the code it sets the same switch as `--no-cache`, so it skips the cache. Do not rely on `--only-cache`. Issue: [#41](https://github.com/awesome-foundation/cfnrain/issues/41).

If you find another difference between this site and the code, please [open an issue](https://github.com/awesome-foundation/cfnrain/issues).
