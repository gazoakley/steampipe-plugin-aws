# something like these scenarios perhaps?

[Assume Role](https://github.com/aws/aws-sdk-go#configuring-credentials)

[Credentials](https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/)

[SSO Creds](https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/ssocreds/)

- using Keys
- using Profiles
- using Env vars
- using assume role (via profile only currently?)
- assume role / MFA (current crappy workaround for now)
- federated / SAML ? Not sure if we have anything for this? Okta? Google? etc

### Using Keys

The AWS plugin allows you set static credentials with the access_key, secret_key, and session_token arguments. You may select one or more regions with the regions argument. An AWS connection may connect to multiple regions, however be aware that performance may be negatively affected by both the number of regions and the latency to them.

```hcl
# credentials via key pair
connection "aws_account_x" {
  plugin      = "aws"
  access_key  = "ASIA3ODZSWFYSN2PFHPJ"
  secret_key  = "gMCYsoGqjfThisISNotARealKeyVVhh"
  regions     = ["us-east-1" , "us-west-2"]
}

connection "aws_account_y" {
  plugin        = "aws"
  access_key    = "ASIA3ODZSWFYSN2PFHPJ"
  secret_key    = "gMCYsoGqjfThisISNotARealKeyVVhh"
  session_token = "IQoJb3JpZ2luX2VjEF8aCXVzLWVhc3QtMSJIMEYCIQC31ra+pa8asV1zF2oRAipT156OP4gG7+7se3IRDSaEIwIhAOcZRvXifmEJ503iU0cPzhQCxsAOHn7i9koMrAfMcKOUKo8CCBgQAhoMMDEzMTIyNTUwOTk2IgxMLJNr3kk1eolERlMq7AEx+QOuiCPORbwsEzGj5a1pQ1grn2QMhhI/UPkgHjwnPHnBe9vd7i9XtPkToNmFYd5gJLj0PaeHKXpgfnjSlg+NGML00Gfvp1CeRn7IggW5Jn9TEwehQC0XcQAmCIaoNlZBYyDwJwmtqsZOGkcIn+nCU6pcjCQGg1dg/vx78WUlTqyBJRixGCdA9YcTkWrNqqDdc/nIxQUxB/kbtW6EZX64gSACBUMn1EwZH79+3y8lPP7pV1W3wpJlgd/E1tBTBQpDr/Y11nL6OlcrDhrpiDdg8SA/hslnCevi4mBzbwFnGZIUeu1NMatu23Gv6zC8+OiFBjqcAVKsmCVREFlhmf8WdxsdcyXyjyAW8gR8nDvcwxkL9iz2OSvDA585uVEjpNZrBuJZHU29K5LFtXu0/xUHMXm5yWLLzw5URioJyGSO/L+P7S5C8RUnP7nnsZY2Elidq+3Fjvsi7vEaQ0d7Sbr8oLgZUtKTQql8znYHWdsKOcK6JL/1m7X4uCoeUn1OnjUDwltSF61tj1kZ/QB1Jn83Mw=="
  regions       = ["us-east-1" , "us-west-2"]
}
```

### Using Profiles

Alternatively, you may select a named profile from an AWS credential file with the profile argument. A connect per profile is a common configuration:

```hcl
# credentials via profile
connection "aws_account_y" {
  plugin      = "aws"
  profile     = "profile_y"
  regions     = ["us-east-1", "us-west-2"]
}

# credentials via profile
connection "aws_account_z" {
  plugin      = "aws"
  profile     = "profile_z"
  regions     = ["us-east-1", "us-west-2"]
}
```

### Using Environment Variables

Credentials specified in environment variables

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_SESSION_TOKEN

- AWS_ROLE_SESSION_NAME

If regions is not specified, Steampipe will use a single default region using the same resolution order as the credentials:

The AWS_DEFAULT_REGION or AWS_REGION environment variable
The region specified in the active profile (AWS_PROFILE or default)

```hcl
# credentials via profile
connection "aws_account_y" {
  plugin      = "aws"
  profile     = "profile_y"
  regions     = ["us-east-1", "us-west-2"]
}

# credentials via profile
connection "aws_account_z" {
  plugin      = "aws"
  profile     = "profile_z"
  regions     = ["us-east-1", "us-west-2"]
}
```

### Using Assume Role/Without MFA - Currently supported through profiles

User must create two AWS profiles:

1. Profile having the details of the `role_arn`, `role_session_name` (if any applicable) and `source_profile` (i.e. The name of the profile haing details for user who will asssume this role.)

2. Source Profile (i.e. AWS profile containing the access credentials for User assuming the role.)

```bash
[profile role-profile]
role_arn = arn:aws:iam::011223344567:role/test_assume
role_session_name = role-without-mfa
source_profile = user-profile

[profile user-profile]
aws_access_key_id = ABCDQGDREFGHCCNHIJKL
aws_secret_access_key = gMCYsoGqjfThisIsNotARealKeyVVhh
region = us-west-2
```

```hcl
connection "role_aws" {
  plugin  = "aws"
  profile = "role-without-mfa"
  regions = ["us-east-1", "us-east-2"]
}
```

### [Using Assume Role/With MFA](https://stackoverflow.com/questions/52432717/terraform-unable-to-assume-roles-with-mfa-enabled/66878739#66878739)

Currently steampipe doesn't support input prompt to allow user to pass mfa token at run time.
To overcome this problem one should generate an AWS profile with temporary credentials.

An way to handle this is to use `credential_process` in order to generate the credentials with a local script and cache the tokens in a new profile (let's call it tf_temp)

This script[(`mfa.sh`)](scripts/mfa.sh) would:

- move this script in `/bin/mfa.sh` or `/usr/local/bin/mfa.sh` and change permission to execute.
- check if the token is still valid for the profile `tf_temp`
- if token is valid, extract the token from existing config using `aws configure get xxx --profile tf_temp`
- if token is not valid, prompt use to enter mfa token
- generate the session token with `aws assume-role --token-code xxxx ... --profile your_profile`
- set the temporary profile token tf_temp using `aws configure set xxx --profile tf_temp`
- before executing below script add below to `~/.aws/credentials` where profile `prod` contains the user access key details who will assume the role.

  ```bash
  [prod]
  aws_secret_access_key = redacted
  aws_access_key_id = redacted

  [tf_temp]

  [tf]
  credential_process = sh -c 'mfa.sh arn:aws:iam::{account_id}:role/{role} arn:aws:iam::{account_id}:mfa/{mfa_entry} prod 2> $(tty)'
  ```

- Add below profile to steampipe connections
  ```hcl
  connection "role_mfa" {
    plugin  = "aws"
    profile = "tf_temp"
    regions = ["us-east-1", "us-east-2"]
  }
  ```
- Now run `aws sts get-caller-identity --profile tf` to generate temporary credentials in `tf_temp` profile.

### [Using AWS SSO](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html#sso-configure-profile-auto)

- You can add an AWS SSO enabled profile to your AWS CLI by running the following command, providing your AWS SSO start URL and the AWS Region that hosts the AWS SSO directory.

  ```shell
  aws configure sso
  SSO start URL [None]: https://my-sso-portal.awsapps.com/start
  SSO region [None]:us-east-1
  ```

- The AWS CLI attempts to open your default browser and begin the login process for your AWS SSO account. If the AWS CLI cannot open the browser, the following message appears with instructions on how to manually start the login process.

  ```shell
  Using a browser, open the following URL:

  https://my-sso-portal.awsapps.com/verify

  and enter the following code:
  QCFK-N451
  ```

- Next, the AWS CLI displays the AWS accounts available for you to use. If you are authorized to use only one account, the AWS CLI selects that account for you automatically and skips the prompt. The AWS accounts that are available for you to use are determined by your user configuration in AWS SSO.

  ```shell
  There are 2 AWS accounts available to you.
  > DeveloperAccount, developer-account-admin@example.com (123456789011)
    ProductionAccount, production-account-admin@example.com (123456789022)
  ```

- Next, the AWS CLI confirms your account choice, and displays the IAM roles that are available to you in the selected account. If the selected account lists only one role, the AWS CLI selects that role for you automatically and skips the prompt. The roles that are available for you to use are determined by your user configuration in AWS SSO.

  ```shell
  Using the account ID 123456789011
  There are 2 roles available to you.
  > ReadOnly
    FullAccess
  ```

- Now you can finish the configuration of your profile, by specifying the default output format, the default AWS Region to send commands to, and providing a name for the profile so you can reference this profile from among all those defined on the local computer.

  ```shell
  CLI default client Region [None]: us-west-2<ENTER>
  CLI default output format [None]: json<ENTER>
  CLI profile name [123456789011_ReadOnly]: my-dev-profile<ENTER>
  ```

- Now you can use this aws profile in your steampipe connection
  ```hcl
  connection "aws_sso" {
    plugin  = "aws"
    profile = "my-dev-profile"
    regions = ["us-east-1", "us-east-2"]
  }
  ```
