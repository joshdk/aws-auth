> ⚠️ This repo has been deprecated in favor of [joshdk/aws-console](https://github.com/joshdk/aws-console).  

---

[![Actions][github-actions-badge]][github-actions-link]
[![License][license-badge]][license-link]
[![Releases][github-release-badge]][github-release-link]

# AWS Auth

🔐 Manage AWS credential for a range of workflows

## Installing

A prebuilt [release][github-release-link] binary can be downloaded by running:

```bash
$ wget -q https://github.com/joshdk/aws-auth/releases/download/v0.1.0/aws-auth-linux-amd64.tar.gz
$ tar -xf aws-auth-linux-amd64.tar.gz
$ sudo install aws-auth /usr/bin/aws-auth
```

Alternatively, a development version of this tool can be installed by running:

```bash
$ go get -u github.com/joshdk/aws-auth
```

## Configuration

### Configs and Credentials

The `aws-auth` tool uses the AWS configuration files (located at `~/.aws/config` and `~/.aws/credentials`) as the source of profile definitions.

For background information on these two file, please take a look at:

- https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html
- https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html
- https://docs.aws.amazon.com/credref/latest/refdocs/file-format.html

### Profiles

A named profile within the AWS config can define a few different things:

- A **user** defines AWS credentials that can be used directly:

```ini
[default]
aws_access_key_id = AKIA...RBDY
aws_secret_access_key = asBY...z6WT
```

- A **role** can be used after an `assume-role` API call. A named profile is also referenced, and is used as the source credentials for the API call.

```ini
[profile dev-role]
source_profile = default
role_arn = arn:aws:iam::000000000000:role/my-role
```

- A **session** can be used after an `get-session-token` API call. A named profile is also referenced, and is used as the source credentials for the API call.

```ini
[profile dev-session]
source_profile = default
```

Note - The session configuration is non-standard and will not work with the AWS CLI.

### Profile Chains

By using a series of named profile references, a profile "chain" can be defined which describes a series of role/session profile that can be used to derive new credentials from an original user.

```ini
[default]
aws_access_key_id = AKIA...RBDY
aws_secret_access_key = asBY...z6WT

[profile temp]
source_profile = default

[profile production]
source_profile = temp
role_arn = arn:aws:iam::000000000000:role/my-role
```

In the above example, we have a `default` profile user which has AWS credentials.

There is also a `temp` profile session, which can derive its credentials from the `default` profile using `get-session-token`.

Finally, there is a `production` profile role, which can derive its credentials from the `temp` profile using `assume-role`.

If credentials for the `production` profile are requested, `aws-auth` will automate the series of necessary API calls.

### Multi-Factor Authentication

You can configure `aws-auth` to prompt for MFA codes if necessary.

```ini
[profile dev-role]
source_profile = default
role_arn = arn:aws:iam::000000000000:role/my-role
mfa_serial = arn:aws:iam::000000000000:mfa/my-user
mfa_message = Please enter MFA code for dev:
```

The `mfa_serial` property references a virtual MFA device that has already been configured for an IAM user in AWS.

The `mfa_message` property can be used to display a custom message to the user.
Note: This property is non-standard and will be ignored by the AWS CLI.

### Yubikeys

If you have enrolled a Yubikey as your MFA device, you can configure `aws-auth` to prompt your Yubikey to generate an MFA code directly.

```ini
[profile dev-role]
source_profile = default
role_arn = arn:aws:iam::000000000000:role/my-role
mfa_serial = arn:aws:iam::000000000000:mfa/my-user
mfa_message = Please touch your Yubikey now!
yubikey_slot = aws-dev-mfa
```

The `yubikey_slot` property can be used to specify the Yubikey oath slot used for generating a code.
Note: This property is non-standard and will be ignored by the AWS CLI.

## Usage

### Help!

```
$ aws-auth --help

aws-auth - Manage AWS credential for a range of workflows

Usage:
  aws-auth [flags]
  aws-auth [command]

Available Commands:
  console     Generate an AWS Console login URL
  help        Help about any command

Flags:
  -h, --help             help for aws-auth
  -p, --profile string   config profile to target (default "default")
  -v, --version          version for aws-auth

Use "aws-auth [command] --help" for more information about a command.
```

### Generating Credentials

Credentials (in the form of `export`-able environment variables) for a named profile can be generated like so:

```shell
$ aws-auth --profile dev

export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_SESSION_TOKEN=...
export AWS_ARN=...
export AWS_ACCOUNT_ID=...
export AWS_EXPIRATION=...
```

### Console Login

A login URL for the AWS Console can also be generated for a role:

```shell
$ aws-auth --profile dev console

https://signin.aws.amazon.com/federation?Action=login...
```

## License

This code is distributed under the [MIT License][license-link], see [LICENSE.txt][license-file] for more information.

[github-actions-badge]:  https://github.com/joshdk/aws-auth/workflows/Build/badge.svg
[github-actions-link]:   https://github.com/joshdk/aws-auth/actions
[github-release-badge]:  https://img.shields.io/github/release/joshdk/aws-auth/all.svg
[github-release-link]:   https://github.com/joshdk/aws-auth/releases
[license-badge]:         https://img.shields.io/badge/license-MIT-green.svg
[license-file]:          https://github.com/joshdk/aws-auth/blob/master/LICENSE.txt
[license-link]:          https://opensource.org/licenses/MIT
