# info

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
