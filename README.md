[![Unit tests](https://github.com/awesome-foundation/cfnrain/actions/workflows/test.yml/badge.svg)](https://github.com/awesome-foundation/cfnrain/actions/workflows/test.yml)

# cfnrain

cfnrain is a command line tool for [AWS CloudFormation](https://aws.amazon.com/cloudformation/) templates and stacks. It deploys stacks through change sets, packages local files into templates, formats templates, and shows stack logs.

cfnrain is a community fork of [AWS CloudFormation Rain](https://github.com/aws-cloudformation/rain). AWS archived Rain on 2026-07-29. cfnrain continues from Rain's last commit, and works as a replacement for Rain v1.24.4.

**cfnrain is not affiliated with or endorsed by Amazon Web Services.**

## Documentation

Read the documentation at **[awesome-foundation.github.io/cfnrain](https://awesome-foundation.github.io/cfnrain/)**. It covers installation, a first deploy, every command, the template directives, modules, CI usage, and migrating from Rain.

## Install

Download an archive for your system from the [releases page](https://github.com/awesome-foundation/cfnrain/releases), and check it against `checksums.txt` and its build provenance:

```sh
sha256sum --ignore-missing -c checksums.txt
gh attestation verify <archive> -R awesome-foundation/cfnrain
```

Or build it with Go:

```sh
go install github.com/awesome-foundation/cfnrain/cmd/cfnrain@latest
```

Coming from Rain? Link the old name, so your scripts keep working:

```sh
ln -s "$(command -v cfnrain)" /usr/local/bin/rain
```

In GitHub Actions, use [install-cfnrain](https://github.com/awesome-foundation/install-cfnrain):

```yaml
- uses: awesome-foundation/install-cfnrain@v1
```

See [Install](https://awesome-foundation.github.io/cfnrain/install.html) and [Migrating from rain](https://awesome-foundation.github.io/cfnrain/migrating.html) for the details.

## Contributing

cfnrain is written in [Go](https://go.dev/). To contribute, [fork this repository](https://github.com/awesome-foundation/cfnrain/fork), make your change, and open a pull request. Questions and ideas go to [GitHub Discussions](https://github.com/awesome-foundation/cfnrain/discussions).

The documentation site is the Markdown in `docs/`. When you add or change a command or a flag, update the page in `docs/` that describes it. A unit test fails when a command has no page. To build the site, install [Zensical](https://zensical.org/) and run `zensical build --clean --strict` in the repository root.

## License

cfnrain is licensed under the Apache 2.0 License. The `NOTICE` file keeps the original Rain copyright.
