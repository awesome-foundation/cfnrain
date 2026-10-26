# Install

cfnrain is one static binary. Download a release archive, or build it with Go.

## Release archives

Each [release](https://github.com/awesome-foundation/cfnrain/releases) has archives for Linux, macOS and Windows:

| OS | Architectures | Format |
|----|---------------|--------|
| Linux | `amd64`, `arm64`, `arm` (v7), `386` | `.tar.gz` |
| macOS | `amd64`, `arm64` | `.tar.gz` |
| Windows | `amd64`, `386` | `.zip` |

The archive name is `cfnrain_<version>_<os>_<arch>`, for example `cfnrain_1.25.2_linux_amd64.tar.gz`. Each archive contains `cfnrain`, `LICENSE`, `NOTICE` and `README.md`. Each release also has a `checksums.txt` file.

Download an archive and the checksum file. This example uses the GitHub CLI:

```sh
VERSION=1.25.2
gh release download "v${VERSION}" -R awesome-foundation/cfnrain \
  -p "cfnrain_${VERSION}_linux_amd64.tar.gz" -p checksums.txt
```

You can also download them from the release page with a browser or `curl`.

### Verify the checksum

`checksums.txt` lists the SHA-256 checksum of every archive. Check the archive that you downloaded:

```sh
sha256sum --ignore-missing -c checksums.txt
```

The output must say `OK` for your archive. On macOS, use `shasum -a 256 --ignore-missing -c checksums.txt`.

### Verify the provenance

The release workflow attests the build provenance of every archive. The attestation proves that the archive was built by the cfnrain release workflow from this repository. Verify it with the GitHub CLI:

```sh
gh attestation verify "cfnrain_${VERSION}_linux_amd64.tar.gz" -R awesome-foundation/cfnrain
```

### Install the binary

```sh
tar -xzf "cfnrain_${VERSION}_linux_amd64.tar.gz" cfnrain
sudo install cfnrain /usr/local/bin/cfnrain
cfnrain --version
```

`cfnrain --version` prints the version, the OS and the architecture, for example `cfnrain v1.25.2 linux/amd64`.

## Install with Go

If you have Go installed, build and install the latest release:

```sh
go install github.com/awesome-foundation/cfnrain/cmd/cfnrain@latest
```

Go puts the binary in `$(go env GOPATH)/bin`. Make sure that this directory is on your `PATH`.

## Homebrew

There is no Homebrew formula yet. Follow [issue #14](https://github.com/awesome-foundation/cfnrain/issues/14) for progress.

## Keep the rain name

Scripts that call `rain` keep working when you link the name to `cfnrain`:

```sh
ln -s "$(command -v cfnrain)" /usr/local/bin/rain
```

You can use any directory on your `PATH`. The shell completion scripts complete both names.

## Shell completion

The repository has completion scripts for bash and zsh:

- [`docs/bash_completion.sh`](https://github.com/awesome-foundation/cfnrain/blob/main/docs/bash_completion.sh)
- [`docs/zsh_completion.sh`](https://github.com/awesome-foundation/cfnrain/blob/main/docs/zsh_completion.sh)

For bash, source the script from your `~/.bashrc`. For zsh, copy the script to a directory on your `fpath` with the name `_cfnrain`.

## AWS credentials

cfnrain uses the standard AWS SDK credential chain: environment variables, the shared config and credentials files, SSO, and instance or container roles. Use `-p` or `--profile` to pick a profile and `-r` or `--region` to pick a region.

Check which account and region cfnrain uses:

```sh
cfnrain info
```
