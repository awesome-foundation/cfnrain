# Packages

A package is a zip file with modules in it. The `Packages` section gives a package a short alias. A module `Source` that starts with `$<alias>/` refers to a file inside that package.

## The Packages section

```yaml
Packages:
  samples:
    Source: https://github.com/awesome-foundation/cfnrain/releases/download/v1.25.2/modules.zip
    Hash: https://github.com/awesome-foundation/cfnrain/releases/download/v1.25.2/modules.sha256

Modules:
  Data:
    Source: $samples/encrypted-bucket.yaml
```

| Key | What it does |
|-----|--------------|
| `Source` | The zip file: a local path, an `https://` URL or an `s3://` URI |
| `Hash` | Optional. A URL or a local path to a file that has the SHA-256 hash of the zip, as hex. If the hash does not match, packaging stops with `hash does not match`. |

A local zip works the same way. This test packages a module from `package.zip`:

The template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/zip-template.yaml"
```

`zip-module.yaml`, inside `package.zip`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/zip-module.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/zip-expect.yaml"
```

## Sample modules

Each cfnrain release has `modules.zip` and `modules.sha256`, with the sample modules from the [`modules`](https://github.com/awesome-foundation/cfnrain/tree/main/modules) directory. Use them as examples for your own modules:

| Module | What it creates |
|--------|-----------------|
| `simple-vpc.yaml` | A VPC with two availability zones |
| `encrypted-bucket.yaml` | A bucket with encryption and blocked public access |
| `compliant-bucket.yaml` | A bucket with log and replica buckets and a bucket policy, made to pass common compliance checks |
| `bucket-policy.yaml` | A bucket policy that denies requests without TLS |
| `load-balancer.yaml` | An Elastic Load Balancing v2 load balancer |
| `static-site.yaml` | An S3 bucket and a CloudFront distribution for a static web site |
| `cognito.yaml` | A Cognito user pool and related resources |
| `rest-api.yaml` | An API Gateway REST API |
| `api-resource.yaml` | A Lambda function and its API Gateway resources |
| `simple-table.yaml` | A DynamoDB table |

## Packages in the Rain section

The `Rain:` section can also define packages, with `Location` instead of `Source`. This is the older form. It works with a resource `Type` that starts with `$<alias>/`. cfnrain turns that type into the older [`!Rain::Module`](#the-older-rainmodule-directive) directive, so it needs `--experimental`.

```yaml
Rain:
  Packages:
    abc:
      Location: ./my-modules
Resources:
  Foo:
    Type: $abc/bucket.yaml
```

## The older !Rain::Module directive

Before the `Modules` section, Rain had module resources: a resource with `Type: !Rain::Module <uri>`. They still work, for compatibility, and need `--experimental`:

```yaml
Resources:
  Show1Hide2:
    Type: !Rain::Module ./ifparam-module.yaml
    Properties:
      Foo: foo
```

Without `--experimental`, packaging stops with `You must add the --experimental arg to use the !Rain::Module directive`. Use the `Modules` section for new templates.

The `pkg` help text in Rain said that a `!Rain::Module` file must have a resource called `ModuleExtension` with `Metadata.Extends`. The code does not check for either of these.

## CodeArtifact modules (experimental)

`cfnrain module` publishes a directory of module files to AWS CodeArtifact as a generic package, and installs it again. Every `module` command needs `-x` (`--experimental`).

```sh
# Create the CodeArtifact domain and repository
cfnrain module bootstrap -x

# Publish the modules in ./modules as version 1.0.0
cfnrain module publish -x my-modules --path modules --version 1.0.0

# Install them into ./vendor/my-modules
cfnrain module install -x my-modules --path vendor/my-modules --version 1.0.0
```

| Flag | Default | What it does |
|------|---------|--------------|
| `--domain` | `cloudformation` | The CodeArtifact domain |
| `--repo` | `rain` | The CodeArtifact repository |
| `--path` | `.` | The local directory to publish from, or to install to |
| `--version` | | The package version. `install` uses the latest version when you do not set it. |

`install` checks the SHA-256 hash of the package against CodeArtifact before it unzips it. After the install, refer to the modules with a local `Source` path.
