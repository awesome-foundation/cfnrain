# Other commands

These commands do not work with templates or stacks directly. Both take `-p` (`--profile`) and `-r` (`--region`).

The output on this page is illustrative. It needs a real AWS account.

## info

`cfnrain info` shows the account, region, identity and profile that cfnrain uses.

```sh
cfnrain info
```

```text
Account:  123456789012
Region:   eu-west-1
Identity: arn:aws:sts::123456789012:assumed-role/Developer/jane
```

`-c` (`--creds`) also prints the credential source and the credentials.

!!! danger "--creds prints your secret key"
    `cfnrain info --creds` prints the `SecretAccessKey` and the `SessionToken` in plain text. Do not run it in CI logs, screen shares or recorded terminals.

Reference: [cfnrain info](../cfnrain_info.md)

## console

`cfnrain console` makes a sign-in URL for the AWS console from your current credentials and opens it in a browser.

```sh
# Open the CloudFormation console
cfnrain console

# Open the CloudFormation console with a stack selected
cfnrain console app-prod

# Open another service
cfnrain console -s lambda

# Print the URL and do not open a browser
cfnrain console -u
```

| Flag | What it does |
|------|--------------|
| `-s`, `--service` | The service home page to open. The default is `cloudformation`. |
| `-u`, `--url` | Print the URL only |
| `-n`, `--name` | The user name to show in the console. The default comes from the role name. |
| `-l`, `--logout` | Sign out of the AWS console |

`console` works only with IAM role credentials, not with IAM user credentials.

The sign-in URL names `https://awesome-foundation.github.io/cfnrain/cfnrain_console.html` as the issuer. That is the [reference page](../cfnrain_console.md) for this command on this site. When your console session expires, AWS links back to it.

Reference: [cfnrain console](../cfnrain_console.md)
