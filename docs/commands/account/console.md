# console

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
| `-s`, `--service` | The service home page to open. The default is `cloudformation`. cfnrain selects the stack only on the `cloudformation` page. |
| `-u`, `--url` | Print the URL only |
| `-n`, `--name` | The user name to show in the console. The default comes from the role name. |
| `-l`, `--logout` | Sign out of the AWS console |

`console` works only with IAM role credentials, not with IAM user credentials.

The sign-in URL names `https://awesome-foundation.github.io/cfnrain/cfnrain_console.html` as the issuer. When your console session expires, AWS links back to [that page](../../cfnrain_console.md).
