## cfnrain module install

Install a package of Rain modules from CodeArtifact

```
cfnrain module install <name> [flags]
```

### Options

```
      --debug            Output debugging information
      --domain string    The CodeArtifact domain (default "cloudformation")
  -x, --experimental     Acknowledge that this is an experimental feature
  -h, --help             help for install
      --path string      The local path for module files, defaults to the current directory (default ".")
  -p, --profile string   AWS profile name; read from the AWS CLI configuration file
  -r, --region string    AWS region to use
      --repo string      The CodeArtifact repository (default "rain")
      --version string   Version of the module to install
```

### Options inherited from parent commands

```
      --no-colour   Disable colour output
```

### SEE ALSO

* [cfnrain module](cfnrain_module.md)	 - Interact with Rain modules in CodeArtifact

